package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/service"
)

type FilesHandler struct {
	svc *service.FilesService
}

func NewFilesHandler(svc *service.FilesService) *FilesHandler {
	return &FilesHandler{svc: svc}
}

func (h *FilesHandler) ListFiles(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	parentSlug := c.QueryParam("parent_slug")
	mimeType := c.QueryParam("mime_type")

	if mimeType != "" {
		items, total, err := h.svc.ListFilesByMime(c.Request().Context(), userID, mimeType, page, pageSize)
		if err != nil {
			return err
		}
		return OK(c, map[string]any{
			"files": items,
			"total": total,
		})
	}

	items, total, err := h.svc.ListFiles(c.Request().Context(), userID, parentSlug, page, pageSize)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"files": items,
		"total": total,
	})
}

func (h *FilesHandler) Mkdir(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		DirName    string `json:"dir_name"`
		ParentSlug string `json:"parent_slug"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	item, err := h.svc.Mkdir(c.Request().Context(), userID, input.DirName, input.ParentSlug)
	if err != nil {
		return err
	}

	return Created(c, item)
}

func (h *FilesHandler) CheckConflict(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		FileName   string `json:"file_name"`
		FileSize   int64  `json:"file_size"`
		PreHash    string `json:"pre_hash"`
		ParentSlug string `json:"parent_slug"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	resp, err := h.svc.CheckConflict(c.Request().Context(), userID, input.FileName, input.PreHash, input.ParentSlug, input.FileSize)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *FilesHandler) CheckDuplicate(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		FileHash   string `json:"file_hash"`
		ParentSlug string `json:"parent_slug"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	resp, err := h.svc.CheckDuplicate(c.Request().Context(), userID, input.FileHash, input.ParentSlug)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *FilesHandler) ImportFile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		PhysicalFileSlug string `json:"physical_file_slug"`
		FileName         string `json:"file_name"`
		ParentSlug       string `json:"parent_slug"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	resp, err := h.svc.ImportFile(c.Request().Context(), userID, input.PhysicalFileSlug, input.FileName, input.ParentSlug)
	if err != nil {
		return err
	}

	return Created(c, resp)
}

func (h *FilesHandler) GetBreadcrumb(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	items, err := h.svc.GetBreadcrumb(c.Request().Context(), userID, slug)
	if err != nil {
		return err
	}

	return OK(c, items)
}

func (h *FilesHandler) TrashFile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	if err := h.svc.TrashFile(c.Request().Context(), userID, slug); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "file trashed"})
}

func (h *FilesHandler) RestoreFile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	if err := h.svc.RestoreFile(c.Request().Context(), userID, slug); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "file restored"})
}

func (h *FilesHandler) PermanentDelete(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	if err := h.svc.PermanentDelete(c.Request().Context(), userID, slug); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "file permanently deleted"})
}

func (h *FilesHandler) RenameFile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	var input struct {
		NewName string `json:"new_name"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	if err := h.svc.RenameFile(c.Request().Context(), userID, slug, input.NewName); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "file renamed"})
}

func (h *FilesHandler) MoveFile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	var input struct {
		TargetParentSlug string `json:"target_parent_slug"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	if err := h.svc.MoveFile(c.Request().Context(), userID, slug, input.TargetParentSlug); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "file moved"})
}

func (h *FilesHandler) SetStarred(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	var input struct {
		Starred bool `json:"starred"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	if err := h.svc.SetStarred(c.Request().Context(), userID, slug, input.Starred); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "star updated"})
}

func (h *FilesHandler) DownloadFile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	file, name, mimeType, err := h.svc.DownloadFile(c.Request().Context(), userID, slug)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, name))
	c.Response().Header().Set(echo.HeaderContentType, mimeType)
	c.Response().WriteHeader(http.StatusOK)

	_, err = io.Copy(c.Response().Writer, file)
	return err
}

func (h *FilesHandler) ListTrashed(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	items, total, err := h.svc.ListTrashed(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"files": items,
		"total": total,
	})
}

func (h *FilesHandler) ListStarred(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	items, total, err := h.svc.ListStarred(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"files": items,
		"total": total,
	})
}
