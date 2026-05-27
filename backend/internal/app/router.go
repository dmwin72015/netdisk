package app

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/config"
	mw "github.com/netdisk/server/internal/middleware"
	"github.com/netdisk/server/pkg/jwtutil"
)

const (
	apiRequestsPerMin  = 600
	authRequestsPerMin = 20
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

func registerRoutes(e *echo.Echo, rdb *redis.Client, jwtMgr *jwtutil.Manager, h *handlers) {
	e.GET("/healthz", healthHandler)

	api := e.Group("/api/v1")
	api.Use(mw.RateLimit(rdb, "api", apiRequestsPerMin, time.Minute))

	// Auth routes
	auth := api.Group("/auth")
	auth.Use(mw.RateLimit(rdb, "auth", authRequestsPerMin, time.Minute))
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)
	auth.POST("/refresh", h.Auth.Refresh)
	auth.POST("/logout", h.Auth.Logout, mw.JWT(jwtMgr))

	// Authenticated routes
	authed := api.Group("", mw.JWT(jwtMgr))

	// User routes
	authed.GET("/user/me", h.User.GetMe)
	authed.PATCH("/user/profile", h.User.UpdateProfile)
	authed.POST("/user/me/password", h.User.ChangePassword)
	authed.POST("/user/me/avatar", h.User.UploadAvatar)
	authed.GET("/user/transactions", h.User.ListTransactions)

	// Public avatar route
	api.GET("/user/avatar/:user_slug", h.User.GetAvatar)

	// File routes
	files := authed.Group("/files")
	files.GET("", h.Files.ListFiles)
	files.POST("/mkdir", h.Files.Mkdir)
	files.POST("/check-conflict", h.Files.CheckConflict)
	files.POST("/check-duplicate", h.Files.CheckDuplicate)
	files.POST("/import", h.Files.ImportFile)
	files.GET("/trash", h.Files.ListTrashed)
	files.GET("/starred", h.Files.ListStarred)
	files.GET("/:slug/breadcrumb", h.Files.GetBreadcrumb)
	files.DELETE("/:slug", h.Files.TrashFile)
	files.POST("/:slug/restore", h.Files.RestoreFile)
	files.DELETE("/:slug/permanent", h.Files.PermanentDelete)
	files.POST("/:slug/rename", h.Files.RenameFile)
	files.POST("/:slug/move", h.Files.MoveFile)
	files.POST("/:slug/star", h.Files.SetStarred)
	files.GET("/:slug/download", h.Files.DownloadFile)

	// Upload routes
	upload := authed.Group("/upload")
	upload.POST("/pre-check", h.Upload.PreCheck)
	upload.POST("/request-challenge", h.Upload.RequestChallenge)
	upload.POST("/verify", h.Upload.Verify)
	upload.POST("/init", h.Upload.Init)
	upload.POST("/chunk", h.Upload.UploadChunk)
	upload.POST("/complete", h.Upload.Complete)
	upload.POST("/update-hash", h.Upload.UpdateHash)
	upload.GET("/:upload_slug/status", h.Upload.GetStatus)

	// Media routes
	media := authed.Group("/media")
	media.POST("/items", h.Media.AddToLibrary)
	media.GET("/items", h.Media.ListMediaItems)
	media.GET("/items/:media_slug", h.Media.GetMediaItem)
	media.DELETE("/items/:media_slug", h.Media.RemoveFromLibrary)
	media.GET("/hls/:media_slug/*", h.Media.ServeHLS)
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
