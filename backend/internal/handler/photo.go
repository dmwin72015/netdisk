package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/middleware"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type PhotoHandler struct {
	svc *service.PhotoService
}

func NewPhotoHandler(svc *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{svc: svc}
}

func (h *PhotoHandler) ListPhotos(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	resp, err := h.svc.ListPhotos(c.Request().Context(), userID, middleware.SessionID(c), page, pageSize)
	if err != nil {
		return BizError(err)
	}

	return OK(c, resp)
}

func (h *PhotoHandler) GetPhotoDetail(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	fileSlug := c.Param("file_slug")
	item, err := h.svc.GetPhotoDetail(c.Request().Context(), userID, middleware.SessionID(c), fileSlug)
	if err != nil {
		return BizError(err)
	}

	return OK(c, item)
}

func (h *PhotoHandler) ServeThumbnail(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	fileSlug := c.Param("file_slug")
	res, err := h.svc.GetThumbnailPath(c.Request().Context(), userID, middleware.SessionID(c), fileSlug)
	if err != nil {
		return BizError(err)
	}

	c.Response().Header().Set(echo.HeaderContentType, "image/jpeg")
	c.Response().Header().Set("Cache-Control", "public, max-age=86400, immutable")
	c.Response().Header().Set("ETag", `"`+res.FileHash+`"`)
	return c.File(res.Path)
}

// Album handlers

func (h *PhotoHandler) CreateAlbum(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req service.AlbumCreateRequest
	if err := c.Bind(&req); err != nil {
		return BizError(model.ErrInvalidInput)
	}

	album, err := h.svc.CreateAlbum(c.Request().Context(), userID, req)
	if err != nil {
		return BizError(err)
	}

	return Created(c, album)
}

func (h *PhotoHandler) ListAlbums(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	resp, err := h.svc.ListAlbums(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return BizError(err)
	}

	return OK(c, resp)
}

func (h *PhotoHandler) GetAlbum(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("album_slug")
	album, err := h.svc.GetAlbum(c.Request().Context(), userID, slug)
	if err != nil {
		return BizError(err)
	}

	return OK(c, album)
}

func (h *PhotoHandler) UpdateAlbum(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req service.AlbumUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BizError(model.ErrInvalidInput)
	}

	slug := c.Param("album_slug")
	album, err := h.svc.UpdateAlbum(c.Request().Context(), userID, slug, req)
	if err != nil {
		return BizError(err)
	}

	return OK(c, album)
}

func (h *PhotoHandler) DeleteAlbum(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("album_slug")
	if err := h.svc.DeleteAlbum(c.Request().Context(), userID, slug); err != nil {
		return BizError(err)
	}

	return OK(c, map[string]string{"message": "album deleted"})
}

func (h *PhotoHandler) AddPhotosToAlbum(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		FileSlugs []string `json:"fileSlugs"`
	}
	if err := c.Bind(&req); err != nil {
		return BizError(model.ErrInvalidInput)
	}

	slug := c.Param("album_slug")
	if err := h.svc.AddPhotosToAlbum(c.Request().Context(), userID, slug, req.FileSlugs); err != nil {
		return BizError(err)
	}

	return OK(c, map[string]string{"message": "photos added"})
}

func (h *PhotoHandler) RemovePhotoFromAlbum(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	albumSlug := c.Param("album_slug")
	fileSlug := c.Param("file_slug")

	if err := h.svc.RemovePhotoFromAlbum(c.Request().Context(), userID, albumSlug, fileSlug); err != nil {
		return BizError(err)
	}

	return OK(c, map[string]string{"message": "photo removed"})
}

func (h *PhotoHandler) ListAlbumPhotos(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	albumSlug := c.Param("album_slug")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	resp, err := h.svc.ListAlbumPhotos(c.Request().Context(), userID, middleware.SessionID(c), albumSlug, page, pageSize)
	if err != nil {
		return BizError(err)
	}

	return OK(c, resp)
}

func (h *PhotoHandler) ListPhotoAlbums(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	fileSlug := c.Param("file_slug")
	albums, err := h.svc.ListPhotoAlbums(c.Request().Context(), userID, fileSlug)
	if err != nil {
		return BizError(err)
	}

	return OK(c, map[string]any{"items": albums})
}
