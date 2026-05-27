package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
	"github.com/netdisk/server/internal/service"
	"github.com/netdisk/server/internal/storage"
	"github.com/netdisk/server/internal/store"
	"github.com/netdisk/server/pkg/jwtutil"
)

type App struct {
	cfg    *config.Config
	logger zerolog.Logger

	pg     *pgxpool.Pool
	rdb    *redis.Client
	jwtMgr *jwtutil.Manager
	worker *media.Worker

	echo *echo.Echo
	srv  *http.Server
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*App, error) {
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
	a.jwtMgr = jwtutil.NewManager(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessTTLMin)*time.Minute,
		time.Duration(cfg.JWT.RefreshTTLHour)*time.Hour,
	)

	handlers := buildHandlers(cfg, logger, queries, pg, rdb, a.jwtMgr)

	// Start media worker
	store := storage.NewLocal(cfg.Storage.Root)
	c := cache.New(rdb)
	a.worker = media.NewWorker(queries, pg, cfg, store, c, logger)

	a.echo = echo.New()
	a.echo.HideBanner = true
	a.echo.HTTPErrorHandler = handler.EchoErrorHandler(logger)

	installMiddleware(a.echo, cfg, logger)
	registerRoutes(a.echo, rdb, a.jwtMgr, handlers)

	a.srv = &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      a.echo,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
	}

	return a, nil
}

func (a *App) Run() error {
	go func() {
		a.logger.Info().Int("port", a.cfg.Server.Port).Msg("server listening")
		if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Start media worker in background
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go a.worker.Start(workerCtx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	a.logger.Info().Msg("shutdown signal received")
	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		a.logger.Error().Err(err).Msg("http server shutdown")
	}
	a.releaseInfra()
	return nil
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
		filepath.Join(cfg.Storage.Root, "tmp"),
		filepath.Join(cfg.Storage.Root, "avatars"),
		filepath.Join(cfg.Storage.Root, "hls"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}

type handlers struct {
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Files  *handler.FilesHandler
	Upload *handler.UploadHandler
	Media  *handler.MediaHandler
}

func buildHandlers(
	cfg *config.Config,
	logger zerolog.Logger,
	queries *sqlc.Queries,
	pg *pgxpool.Pool,
	rdb *redis.Client,
	jwtMgr *jwtutil.Manager,
) *handlers {
	authSvc := service.NewAuthService(queries, pg, jwtMgr, cfg)
	userSvc := service.NewUserService(queries, pg, cfg)

	store := storage.NewLocal(cfg.Storage.Root)
	c := cache.New(rdb)

	filesSvc := service.NewFilesService(queries, pg, cfg, store)
	uploadSvc := service.NewUploadService(queries, pg, cfg, store, c)
	mediaSvc := service.NewMediaService(queries, pg, cfg, store, c)

	return &handlers{
		Auth:   handler.NewAuthHandler(authSvc),
		User:   handler.NewUserHandler(userSvc),
		Files:  handler.NewFilesHandler(filesSvc),
		Upload: handler.NewUploadHandler(uploadSvc),
		Media:  handler.NewMediaHandler(mediaSvc),
	}
}
