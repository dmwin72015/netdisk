package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/model"
)

type ConfigHandler struct {
	cfg *config.Config
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

func (h *ConfigHandler) Get(c echo.Context) error {
	device := model.DeviceWeb
	if d := c.QueryParam("device"); d != "" {
		device = model.DeviceType(d)
	}

	resp := model.ClientConfig{
		Device: device,
		Configs: map[string]any{
			"upload.chunkSize":     h.cfg.Upload.ChunkSize,
			"upload.maxUploadSize": h.cfg.Storage.MaxUploadSize,
			"avatar.maxSize":       h.cfg.Limits.AvatarMaxSize,
		},
	}

	return OK(c, resp)
}
