package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/cache"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/handler"
	"github.com/netdisk/server/internal/media"
	"github.com/netdisk/server/internal/pkg/iplookup"
	"github.com/netdisk/server/internal/service"
	"github.com/netdisk/server/internal/storage"
	"github.com/netdisk/server/internal/store"
	"github.com/netdisk/server/pkg/jwtutil"
)

type App struct {
	cfg    *config.Config
	logger zerolog.Logger

	pg          *pgxpool.Pool
	rdb         *redis.Client
	jwtMgr      *jwtutil.Manager
	worker      *media.Worker
	trashWorker *service.TrashWorker

	echo *echo.Echo
	srv  *http.Server
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*App, error) {
	if err := validateStartupConfig(cfg); err != nil {
		return nil, err
	}

	a := &App{cfg: cfg, logger: logger}

	pg, err := store.NewPostgresPool(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	a.pg = pg

	rdb, err := cache.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		a.closePartial()
		return nil, fmt.Errorf("redis: %w", err)
	}
	a.rdb = rdb

	if err := ensureStorageDirs(cfg); err != nil {
		a.closePartial()
		return nil, err
	}

	queries := sqlc.New(pg)

	if err := initSystemConfig(ctx, cfg, pg, logger); err != nil {
		a.closePartial()
		return nil, fmt.Errorf("init system config: %w", err)
	}

	if err := migrateStoragePaths(ctx, cfg, pg); err != nil {
		a.closePartial()
		return nil, fmt.Errorf("migrate storage paths: %w", err)
	}

	if err := recoverStuckDownloadTasks(ctx, pg, logger); err != nil {
		a.closePartial()
		return nil, fmt.Errorf("recover stuck tasks: %w", err)
	}
	a.jwtMgr = jwtutil.NewManager(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessTTLMin)*time.Minute,
		time.Duration(cfg.JWT.RefreshTTLHour)*time.Hour,
	)

	broadcaster := media.NewBroadcaster()
	handlers := buildHandlers(ctx, cfg, logger, queries, pg, rdb, a.jwtMgr, broadcaster)

	// Start media worker
	store := storage.NewLocal(cfg.Storage.Root, cfg.Storage.TmpDir, cfg.Storage.FilesDir)
	c := cache.New(rdb, cfg)
	a.worker = media.NewWorker(queries, pg, cfg, store, c, logger, broadcaster)
	a.trashWorker = service.NewTrashWorker(queries, pg, store, logger, cfg)

	a.echo = echo.New()
	a.echo.HideBanner = true
	a.echo.HTTPErrorHandler = handler.EchoErrorHandler(logger)

	installMiddleware(a.echo, cfg, logger)
	registerRoutes(a.echo, rdb, a.jwtMgr, handlers, cfg, queries)

	a.srv = &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      a.echo,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
	}

	return a, nil
}

func (a *App) Run() error {
	listener, err := net.Listen("tcp", a.srv.Addr)
	if err != nil {
		a.releaseInfra()
		return fmt.Errorf("server listen %s: %w", a.srv.Addr, err)
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	workerErrCh := make(chan error, 2)
	a.startCriticalWorker(workerCtx, workerErrCh, "media worker", a.worker.Start)
	a.startCriticalWorker(workerCtx, workerErrCh, "trash worker", a.trashWorker.Start)

	serverErrCh := make(chan error, 1)
	go func() {
		if err := a.srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()
	a.logger.Info().Int("port", a.cfg.Server.Port).Msg("server listening")
	if err := writeStartupReadyFile(); err != nil {
		workerCancel()
		_ = a.srv.Close()
		a.releaseInfra()
		return fmt.Errorf("write startup ready file: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-serverErrCh:
		workerCancel()
		a.releaseInfra()
		return fmt.Errorf("server failed: %w", err)
	case err := <-workerErrCh:
		workerCancel()
		_ = a.srv.Close()
		a.releaseInfra()
		return fmt.Errorf("critical worker failed: %w", err)
	case sig := <-sigCh:
		a.logger.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	}

	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		a.logger.Error().Err(err).Msg("http server shutdown")
		a.releaseInfra()
		return fmt.Errorf("http server shutdown: %w", err)
	}
	a.releaseInfra()
	return nil
}

func (a *App) startCriticalWorker(ctx context.Context, errCh chan<- error, name string, start func(context.Context)) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				reportCriticalWorkerError(errCh, fmt.Errorf("%s panic: %v", name, recovered))
			}
		}()

		start(ctx)
		if ctx.Err() == nil {
			reportCriticalWorkerError(errCh, fmt.Errorf("%s stopped unexpectedly", name))
		}
	}()
}

func reportCriticalWorkerError(errCh chan<- error, err error) {
	select {
	case errCh <- err:
	default:
	}
}

func validateStartupConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", cfg.Server.Port)
	}
	if cfg.Media.PollInterval <= 0 {
		return fmt.Errorf("media.poll_interval must be greater than 0")
	}
	if cfg.Trash.PollInterval <= 0 {
		return fmt.Errorf("trash.poll_interval must be greater than 0")
	}
	return nil
}

const startupReadyFileEnv = "NETDISK_READY_FILE"

func writeStartupReadyFile() error {
	path := strings.TrimSpace(os.Getenv(startupReadyFileEnv))
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir ready file dir: %w", err)
	}
	return os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339Nano)+"\n"), 0o644)
}

func (a *App) closePartial() {
	a.releaseInfra()
}

func (a *App) releaseInfra() {
	if a.rdb != nil {
		_ = a.rdb.Close()
		a.rdb = nil
	}
	if a.pg != nil {
		a.pg.Close()
		a.pg = nil
	}
}

func ensureStorageDirs(cfg *config.Config) error {
	dirs := []string{
		filepath.Join(cfg.Storage.Root, cfg.Storage.TmpDir),
		filepath.Join(cfg.Storage.Root, cfg.Storage.FilesDir),
		filepath.Join(cfg.Storage.Root, cfg.Storage.AvatarsDir),
		filepath.Join(cfg.Storage.Root, cfg.Storage.HLSDir),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}

var hexDirRe = regexp.MustCompile(`^[a-f0-9]{2}$`)

// migrateStoragePaths moves files from old layout (data/ab/cd/hash) to
// new layout (data/files/ab/cd/hash) and updates DB records.
func migrateStoragePaths(ctx context.Context, cfg *config.Config, pg *pgxpool.Pool) error {
	entries, err := os.ReadDir(cfg.Storage.Root)
	if err != nil {
		return fmt.Errorf("read storage root: %w", err)
	}

	filesRoot := filepath.Join(cfg.Storage.Root, cfg.Storage.FilesDir)
	moved := 0

	for _, e := range entries {
		if !e.IsDir() || !hexDirRe.MatchString(e.Name()) {
			continue
		}

		srcDir := filepath.Join(cfg.Storage.Root, e.Name())
		dstDir := filepath.Join(filesRoot, e.Name())

		if err := os.Rename(srcDir, dstDir); err != nil {
			return fmt.Errorf("move %s -> %s: %w", srcDir, dstDir, err)
		}
		moved++
	}

	if moved > 0 {
		prefix := cfg.Storage.FilesDir + "/"
		tag, err := pg.Exec(ctx,
			`UPDATE physical_files SET storage_path = $1 || storage_path WHERE storage_path NOT LIKE $2`,
			prefix, prefix+"%",
		)
		if err != nil {
			return fmt.Errorf("update db paths: %w", err)
		}
		fmt.Printf("migrated %d dirs, updated %d db rows\n", moved, tag.RowsAffected())
	}

	return nil
}

func initSystemConfig(ctx context.Context, cfg *config.Config, pg *pgxpool.Pool, logger zerolog.Logger) error {
	defs := []struct {
		key   string
		value any
	}{
		{"max_upload_size", cfg.Storage.MaxUploadSize},
		{"chunk_size", cfg.Upload.ChunkSize},
		{"default_quota", cfg.Limits.DefaultStorageQuota},
		{"avatar_max_size", cfg.Limits.AvatarMaxSize},
		{"retention_days", cfg.Trash.RetentionDays},
		{"access_ttl_min", cfg.JWT.AccessTTLMin},
		{"refresh_ttl_hour", cfg.JWT.RefreshTTLHour},
		{"task_expiry_days", cfg.Upload.TaskExpiryDays},
		{"api_requests_per_min", cfg.RateLimit.APIRequestsPerMin},
	}

	for _, d := range defs {
		data, err := json.Marshal(d.value)
		if err != nil {
			return fmt.Errorf("marshal default %s: %w", d.key, err)
		}
		_, err = pg.Exec(ctx, `
			INSERT INTO system_configs (key, value, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (key) DO NOTHING
		`, d.key, data)
		if err != nil {
			return fmt.Errorf("init config %s: %w", d.key, err)
		}
	}
	return nil
}

type handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Files        *handler.FilesHandler
	Share        *handler.ShareHandler
	Upload       *handler.UploadHandler
	Media        *handler.MediaHandler
	Photo        *handler.PhotoHandler
	Config       *handler.ConfigHandler
	Admin        *handler.AdminHandler
	ActivityLog  *handler.ActivityLogHandler
}

func buildHandlers(
	ctx context.Context,
	cfg *config.Config,
	logger zerolog.Logger,
	queries *sqlc.Queries,
	pg *pgxpool.Pool,
	rdb *redis.Client,
	jwtMgr *jwtutil.Manager,
	broadcaster *media.Broadcaster,
) *handlers {
	configSvc := service.NewSystemConfigService(pg, logger)
	if err := configSvc.Load(ctx); err != nil {
		logger.Warn().Err(err).Msg("system config load (using defaults)")
	}

	ipLookup := newIPLookup(cfg.IPLookup, logger)
	auditSvc := service.NewAuditService(pg, ipLookup, logger)

	authSvc := service.NewAuthService(queries, pg, jwtMgr, cfg, rdb, configSvc)
	userSvc := service.NewUserService(queries, pg, cfg)
	adminSvc := service.NewAdminService(queries, pg, logger, cfg.Storage.Root, cfg.Storage.FilesDir, configSvc)

	store := storage.NewLocal(cfg.Storage.Root, cfg.Storage.TmpDir, cfg.Storage.FilesDir)
	c := cache.New(rdb, cfg)

	filesSvc := service.NewFilesService(queries, pg, cfg, store)
	shareSvc := service.NewShareService(pg, store)
	uploadSvc := service.NewUploadService(queries, pg, cfg, store, c, logger)
	mediaSvc := service.NewMediaService(queries, pg, cfg, store, c, filesSvc, logger)
	photoSvc := service.NewPhotoService(queries, pg, cfg, store, logger)

	return &handlers{
		Auth:   handler.NewAuthHandler(authSvc, auditSvc),
		User:   handler.NewUserHandler(userSvc, auditSvc),
		Files:  handler.NewFilesHandler(filesSvc, auditSvc),
		Share:  handler.NewShareHandler(shareSvc, auditSvc),
		Upload: handler.NewUploadHandler(uploadSvc, logger, auditSvc),
		Media:  handler.NewMediaHandler(mediaSvc, broadcaster),
		Photo:  handler.NewPhotoHandler(photoSvc),
		Config:      handler.NewConfigHandler(cfg),
		Admin:       handler.NewAdminHandler(adminSvc, cfg, configSvc, auditSvc),
		ActivityLog: handler.NewActivityLogHandler(queries),
	}
}

// recoverStuckDownloadTasks resets URL download tasks stuck in "downloading"
// status to "failed" on server startup. These tasks get stuck when the server
// restarts or crashes while a background download goroutine is running.
func recoverStuckDownloadTasks(ctx context.Context, pg *pgxpool.Pool, logger zerolog.Logger) error {
	tag, err := pg.Exec(ctx,
		`UPDATE upload_tasks
		 SET status = 'failed', error_msg = 'server restart: download interrupted'
		 WHERE task_type = 'url' AND status = 'downloading'`,
	)
	if err != nil {
		return fmt.Errorf("recover stuck download tasks: %w", err)
	}
	if n := tag.RowsAffected(); n > 0 {
		logger.Warn().Int64("count", n).Msg("recovered stuck URL download tasks")
	}
	return nil
}

func newIPLookup(cfg config.IPLookupConfig, logger zerolog.Logger) iplookup.Lookup {
	switch cfg.Provider {
	case "maxmind":
		lookup, err := iplookup.NewMaxMindLookup(cfg.MaxMindDBPath)
		if err != nil {
			logger.Warn().Err(err).Str("path", cfg.MaxMindDBPath).Msg("maxmind db not available, falling back to http")
			return iplookup.NewHTTPLookup("http://ip-api.com/json/{ip}", "")
		}
		logger.Info().Str("path", cfg.MaxMindDBPath).Msg("ip lookup: maxmind loaded")
		return lookup
	case "http":
		logger.Info().Str("endpoint", cfg.HTTPEndpoint).Msg("ip lookup: http api")
		return iplookup.NewHTTPLookup(cfg.HTTPEndpoint, cfg.HTTPAPIKey)
	default:
		logger.Info().Msg("ip lookup: using free ip-api.com")
		return iplookup.NewHTTPLookup("http://ip-api.com/json/{ip}", "")
	}
}
