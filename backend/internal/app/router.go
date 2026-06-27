package app

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	mw "github.com/netdisk/server/internal/middleware"
	"github.com/netdisk/server/internal/storage"
	"github.com/netdisk/server/internal/web"
	"github.com/netdisk/server/pkg/jwtutil"
)

func installMiddleware(e *echo.Echo, cfg *config.Config, logger zerolog.Logger) {
	e.Use(echomw.Recover())
	e.Use(mw.RequestLogger(logger))
	if len(cfg.Server.CORSOrigins) > 0 {
		e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
			AllowOrigins:     cfg.Server.CORSOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
			AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
			AllowCredentials: true,
		}))
	}
}

func registerRoutes(e *echo.Echo, rdb *redis.Client, jwtMgr *jwtutil.Manager, h *handlers, cfg *config.Config, queries *sqlc.Queries) {
	e.GET("/healthz", healthHandler)

	api := e.Group("/api/v1")
	api.Use(mw.RateLimit(rdb, "api", cfg.RateLimit.APIRequestsPerMin, time.Minute))

	// Auth routes
	auth := api.Group("/auth")
	auth.Use(mw.RateLimit(rdb, "auth", cfg.RateLimit.AuthRequestsPerMin, time.Minute))
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)
	auth.POST("/refresh", h.Auth.Refresh)
	auth.POST("/logout", h.Auth.Logout, mw.JWT(jwtMgr))
	auth.GET("/oauth/:provider/authorize", h.Auth.OAuthRedirect)
	auth.GET("/oauth/:provider/callback", h.Auth.OAuthCallback)

	// OAuth bind (requires auth)
	auth.GET("/oauth/:provider/bind", h.Auth.OAuthBind, mw.JWT(jwtMgr))
	// OAuth bind replace-confirm: account page POSTs here after the user
	// approves replacing their current binding for the same provider.
	auth.POST("/oauth/bind/confirm-replace", h.Auth.OAuthBindConfirmReplace, mw.JWT(jwtMgr))
	// OAuth email-confirm: login page POSTs here after the user confirms
	// linking an OAuth account to an existing local user matched by email.
	auth.POST("/oauth/email-confirm", h.Auth.OAuthEmailConfirm)

	// Public share routes
	publicShares := api.Group("/public/shares")
	publicShares.GET("/:slug", h.Share.GetPublicShare)
	publicShares.POST("/:slug/verify", h.Share.VerifyPublicShare)
	publicShares.GET("/:slug/file", h.Share.ServePublicFile)

	// Authenticated routes
	authed := api.Group("", mw.JWT(jwtMgr))

	// User routes
	authed.GET("/user/me", h.User.GetMe)
	authed.GET("/user/storage-breakdown", h.User.GetStorageBreakdown)
	authed.PATCH("/user/profile", h.User.UpdateProfile)
	authed.POST("/user/me/password", h.User.ChangePassword)
	authed.DELETE("/user/oauth/:provider", h.Auth.OAuthUnlink)
	authed.POST("/user/me/avatar", h.User.UploadAvatar)
	authed.GET("/user/settings", h.User.GetSettings)
	authed.PUT("/user/settings", h.User.UpdateSettings)
	authed.GET("/user/transactions", h.User.ListTransactions)

	// Security log routes
	authed.GET("/user/security-logs", h.ActivityLog.ListSecurityLogs)
	authed.GET("/user/login-devices", h.ActivityLog.ListLoginDevices)

	// Static avatar serving
	avatarDir := filepath.Join(cfg.Storage.Root, cfg.Storage.AvatarsDir)
	api.GET("/avatars/*", func(c echo.Context) error {
		rel := storage.SafePath(c.Param("*"))
		if rel == "" {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Response().Header().Set("Cache-Control", "public, max-age=86400, immutable")
		return c.File(filepath.Join(avatarDir, rel))
	})

	// File routes
	files := authed.Group("/files")
	files.GET("", h.Files.ListFiles)
	files.GET("/recent", h.Files.ListRecentFiles)
	files.POST("/mkdir", h.Files.Mkdir)
	files.POST("/check-conflict", h.Files.CheckConflict)
	files.POST("/check-duplicate", h.Files.CheckDuplicate)
	files.POST("/import", h.Files.ImportFile)
	files.GET("/trash", h.Files.ListTrashed)
	files.POST("/trash/batch", h.Files.BatchTrashFiles)
	files.POST("/trash/empty", h.Files.EmptyTrash)
	files.POST("/trash/restore-all", h.Files.RestoreAll)
	files.GET("/starred", h.Files.ListStarred)
	files.GET("/:slug/breadcrumb", h.Files.GetBreadcrumb)
	files.DELETE("/:slug", h.Files.TrashFile)
	files.POST("/:slug/restore", h.Files.RestoreFile)
	files.DELETE("/:slug/permanent", h.Files.PermanentDelete)
	files.DELETE("/:slug/force", h.Files.ForceDeleteDir)
	files.POST("/:slug/rename", h.Files.RenameFile)
	files.POST("/:slug/move", h.Files.MoveFile)
	files.POST("/:slug/star", h.Files.SetStarred)
	files.POST("/:slug/lock", h.Files.SetDirectoryLock)
	files.DELETE("/:slug/lock", h.Files.ClearDirectoryLock)
	files.POST("/:slug/unlock", h.Files.UnlockDirectory)
	files.GET("/:slug/download", h.Files.DownloadFile)
	files.GET("/:slug/summary", h.Files.GetFolderSummary)

	// Share routes
	shares := authed.Group("/shares")
	shares.POST("", h.Share.CreateShare)
	shares.GET("", h.Share.ListShares)
	shares.PATCH("/:slug", h.Share.UpdateShare)
	shares.DELETE("/:slug", h.Share.CancelShare)

	// Upload routes
	upload := authed.Group("/upload")
	upload.POST("/from-url", h.Upload.UploadFromURL)
	upload.POST("/pre-check", h.Upload.PreCheck)
	upload.POST("/dedup-by-hash", h.Upload.CheckFileDedup)
	upload.POST("/request-challenge", h.Upload.RequestChallenge)
	upload.POST("/verify", h.Upload.Verify)
	upload.POST("/init", h.Upload.Init)
	upload.POST("/chunk", h.Upload.UploadChunk)
	upload.POST("/complete", h.Upload.Complete)
	upload.POST("/update-hash", h.Upload.UpdateHash)
	upload.GET("/tasks", h.Upload.ListTasks)
	upload.POST("/tasks/:slug/retry", h.Upload.RetryTask)
	upload.DELETE("/tasks", h.Upload.DeleteTasks)
	upload.DELETE("/tasks/:slug", h.Upload.DeleteTask)
	upload.GET("/:upload_slug/status", h.Upload.GetStatus)

	// Config route
	authed.GET("/config", h.Config.Get)

	// Admin routes
	admin := authed.Group("/admin", mw.AdminRequired(queries))
	admin.GET("/dashboard/stats", h.Admin.DashboardStats)
	admin.GET("/users", h.Admin.ListUsers)
	admin.POST("/users", h.Admin.CreateUser)
	admin.GET("/users/:id", h.Admin.GetUser)
	admin.PATCH("/users/:id", h.Admin.UpdateUser)
	admin.PATCH("/users/:id/storage-base", h.Admin.UpdateStorageBase)
	admin.DELETE("/users/:id", h.Admin.DeleteUser)
	admin.GET("/files", h.Admin.ListFiles)
	admin.DELETE("/files/:id", h.Admin.DeleteFile)
	admin.PATCH("/files/:id/restore", h.Admin.RestoreFile)
	admin.GET("/storage/stats", h.Admin.StorageStats)
	admin.GET("/system/info", h.Admin.SystemInfo)
	admin.GET("/system/config", h.Admin.ListSystemConfig)
	admin.PUT("/system/config", h.Admin.UpdateSystemConfig)
	admin.POST("/system/config/reset", h.Admin.ResetSystemConfig)
	admin.POST("/cleanup/query", h.Admin.CleanupQuery)
	admin.POST("/cleanup/delete-user-file", h.Admin.CleanupDeleteUserFile)
	admin.POST("/cleanup/delete-physical-file", h.Admin.CleanupDeletePhysicalFile)

	// Photo routes
	photos := authed.Group("/photos")
	photos.GET("", h.Photo.ListPhotos)
	photos.GET("/:file_slug", h.Photo.GetPhotoDetail)
	photos.GET("/:file_slug/thumbnail", h.Photo.ServeThumbnail)
	photos.GET("/:file_slug/albums", h.Photo.ListPhotoAlbums)

	// Album routes
	albums := authed.Group("/albums")
	albums.POST("", h.Photo.CreateAlbum)
	albums.GET("", h.Photo.ListAlbums)
	albums.GET("/:album_slug", h.Photo.GetAlbum)
	albums.PUT("/:album_slug", h.Photo.UpdateAlbum)
	albums.DELETE("/:album_slug", h.Photo.DeleteAlbum)
	albums.POST("/:album_slug/photos", h.Photo.AddPhotosToAlbum)
	albums.GET("/:album_slug/photos", h.Photo.ListAlbumPhotos)
	albums.DELETE("/:album_slug/photos/:file_slug", h.Photo.RemovePhotoFromAlbum)

	// Media routes
	media := authed.Group("/media")
	media.GET("/events", h.Media.StreamEvents)
	media.GET("/upload-dir", h.Media.EnsureUploadDir)
	media.POST("/items", h.Media.AddToLibrary)
	media.GET("/items", h.Media.ListMediaItems)
	media.POST("/items/readd-existing", h.Media.ReaddExistingUpload)
	media.POST("/items/batch-delete", h.Media.BatchRemoveFromLibrary)
	media.GET("/items/:media_slug", h.Media.GetMediaItem)
	media.POST("/items/:media_slug/rename", h.Media.RenameMediaItem)
	media.DELETE("/items/:media_slug", h.Media.RemoveFromLibrary)
	media.GET("/poster/:media_slug", h.Media.ServePoster)
	media.GET("/hls/:media_slug/*", h.Media.ServeHLS)

	serveFrontend(e)
}

func serveFrontend(e *echo.Echo) {
	fileServer := http.FileServer(http.FS(web.BuildFS))
	e.GET("/*", func(c echo.Context) error {
		path := strings.TrimPrefix(c.Request().URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(web.BuildFS, path); err != nil {
				if !shouldServeFrontendFallback(path) {
					return c.NoContent(http.StatusNotFound)
				}
				c.Request().URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

func shouldServeFrontendFallback(path string) bool {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return true
	}
	if path == "api" || strings.HasPrefix(path, "api/") {
		return false
	}
	if path == "_app" || strings.HasPrefix(path, "_app/") || path == "app" || strings.HasPrefix(path, "app/") {
		return false
	}
	lastSegment := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		lastSegment = path[i+1:]
	}
	return !strings.Contains(lastSegment, ".")
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
