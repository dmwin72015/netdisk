package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/media"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type MediaHandler struct {
	svc         *service.MediaService
	broadcaster *media.Broadcaster
}

func NewMediaHandler(svc *service.MediaService, broadcaster *media.Broadcaster) *MediaHandler {
	return &MediaHandler{svc: svc, broadcaster: broadcaster}
}

func (h *MediaHandler) AddToLibrary(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.AddToLibraryRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.AddToLibrary(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return Created(c, resp)
}

func (h *MediaHandler) ReaddExistingUpload(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.ReaddExistingUploadRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.ReaddExistingUpload(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return Created(c, resp)
}

func (h *MediaHandler) EnsureUploadDir(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	item, err := h.svc.EnsureUploadDir(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return OK(c, item)
}

func (h *MediaHandler) ListMediaItems(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

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

func (h *MediaHandler) BatchRemoveFromLibrary(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		MediaSlugs []string `json:"mediaSlugs"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	if err := h.svc.BatchRemoveFromLibrary(c.Request().Context(), userID, input.MediaSlugs); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "removed from library"})
}

func (h *MediaHandler) RenameMediaItem(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		NewName string `json:"newName"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	item, err := h.svc.RenameMediaItem(c.Request().Context(), userID, c.Param("media_slug"), input.NewName)
	if err != nil {
		return err
	}

	return OK(c, item)
}

func (h *MediaHandler) ServePoster(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	mediaSlug := c.Param("media_slug")
	filePath, err := h.svc.GetPosterPath(c.Request().Context(), userID, mediaSlug)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, "image/jpeg")
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	return c.File(filePath)
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

func (h *MediaHandler) StreamEvents(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
	}

	ch := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(ch)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case event := <-ch:
			if event.UserID != userID {
				continue
			}
			data, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(c.Response().Writer, "event: %s\ndata: %s\n\n", event.Status, data)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(c.Response().Writer, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
