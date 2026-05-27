package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/service"
)

type MediaHandler struct {
	svc *service.MediaService
}

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) AddToLibrary(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.AddToLibraryRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	resp, err := h.svc.AddToLibrary(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return Created(c, resp)
}

func (h *MediaHandler) ListMediaItems(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	items, total, err := h.svc.ListMediaItems(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"items": items,
		"total": total,
	})
}

func (h *MediaHandler) GetMediaItem(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	mediaSlug := c.Param("media_slug")
	item, err := h.svc.GetMediaItem(c.Request().Context(), userID, mediaSlug)
	if err != nil {
		return err
	}

	return OK(c, item)
}

func (h *MediaHandler) RemoveFromLibrary(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	mediaSlug := c.Param("media_slug")
	if err := h.svc.RemoveFromLibrary(c.Request().Context(), userID, mediaSlug); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "removed from library"})
}

func (h *MediaHandler) ServeHLS(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	mediaSlug := c.Param("media_slug")
	subPath := c.Param("*")
	if subPath == "" {
		subPath = "index.m3u8"
	}

	filePath, err := h.svc.GetHLSPath(c.Request().Context(), userID, mediaSlug, subPath)
	if err != nil {
		return err
	}

	// Set content type based on extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(subPath, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	} else if strings.HasSuffix(subPath, ".ts") {
		contentType = "video/mp2t"
	}

	c.Response().Header().Set(echo.HeaderContentType, contentType)
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	c.Response().WriteHeader(http.StatusOK)

	return c.File(filePath)
}
