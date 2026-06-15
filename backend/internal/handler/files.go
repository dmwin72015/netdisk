package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/db"
	"github.com/netdisk/server/internal/model"
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
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	params := db.ListFilesParams{
		UserID:      userID,
		Page:        page,
		PageSize:    pageSize,
		IncludeDirs: true,
		// SortBy/SortDir left empty → normalize() defaults to created_at DESC
	}

	if v := c.QueryParam("sortBy"); v != "" {
		params.SortBy = v
	}
	if v := c.QueryParam("sortDir"); v != "" {
		params.SortDir = v
	}
	if v := c.QueryParam("onlyDirs"); v == "true" {
		params.OnlyDirs = true
		params.IncludeDirs = true
	}
	if v := c.QueryParam("includeSystem"); v == "false" {
		params.ExcludeSystem = true
	}

	if parentSlug := c.QueryParam("parentSlug"); parentSlug != "" {
		parent, pErr := h.svc.ResolveParent(c.Request().Context(), userID, parentSlug)
		if pErr != nil {
			return pErr
		}
		params.ParentID = &parent.ID
	}

	if v := c.QueryParam("mimeType"); v != "" {
		params.MimePrefix = &v
		params.IncludeDirs = false
		params.IgnoreParentID = true
	}

	if v := c.QueryParam("fileCategory"); v != "" {
		if v == "folder" {
			params.OnlyDirs = true
			params.IncludeDirs = true
		} else {
			params.Category = &v
			params.IncludeDirs = false
		}
		params.IgnoreParentID = true
	}

	if v := c.QueryParam("searchQuery"); v != "" {
		params.SearchQuery = &v
	} else if v := c.QueryParam("q"); v != "" {
		params.SearchQuery = &v
	}

	items, total, err := h.svc.ListUserFiles(c.Request().Context(), params)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"files": items,
		"total": total,
	})
}

func (h *FilesHandler) ListRecentFiles(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	items, total, err := h.svc.ListRecentFiles(c.Request().Context(), userID, limit)
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
		DirName    string `json:"dirName"`
		ParentSlug string `json:"parentSlug"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
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
		FileName   string `json:"fileName"`
		FileSize   int64  `json:"fileSize"`
		PreHash    string `json:"preHash"`
		ParentSlug string `json:"parentSlug"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
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
		FileHash   string `json:"fileHash"`
		ParentSlug string `json:"parentSlug"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
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
		PhysicalFileSlug string `json:"physicalFileSlug"`
		FileName         string `json:"fileName"`
		ParentSlug       string `json:"parentSlug"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
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

func (h *FilesHandler) BatchTrashFiles(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		Slugs []string `json:"slugs"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}
	if len(input.Slugs) == 0 {
		return model.ErrInvalidInput
	}

	if err := h.svc.BatchTrashFiles(c.Request().Context(), userID, input.Slugs); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "files trashed"})
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
		NewName string `json:"newName"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
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
		TargetParentSlug string `json:"targetParentSlug"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
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
		return model.ErrInvalidInput
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
	res, err := h.svc.DownloadFile(c.Request().Context(), userID, slug)
	if err != nil {
		return err
	}
	defer res.File.Close()

	stat, err := res.File.Stat()
	if err != nil {
		return fmt.Errorf("stat download file: %w", err)
	}

	resp := c.Response()
	resp.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, res.Name))
	resp.Header().Set(echo.HeaderContentType, res.MimeType)
	resp.Header().Set("Cache-Control", "private, max-age=3600")
	resp.Header().Set("ETag", `"`+res.FileHash+`"`)

	http.ServeContent(resp, c.Request(), res.Name, stat.ModTime(), res.File)
	return nil
}

func (h *FilesHandler) ListTrashed(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	params := db.ListFilesParams{
		UserID:      userID,
		IsTrashed:   true,
		IncludeDirs: true,
		SortBy:      "trashed_at",
		SortDir:     "DESC",
		Page:        page,
		PageSize:    pageSize,
	}

	items, total, err := h.svc.ListUserFiles(c.Request().Context(), params)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"files": items,
		"total": total,
	})
}

func (h *FilesHandler) EmptyTrash(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	count, err := h.svc.EmptyTrash(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"deleted": count,
	})
}

func (h *FilesHandler) RestoreAll(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	count, err := h.svc.RestoreAll(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"restored": count,
	})
}

func (h *FilesHandler) ListStarred(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	starred := true
	params := db.ListFilesParams{
		UserID:         userID,
		IsStarred:      &starred,
		IncludeDirs:    true,
		IgnoreParentID: true,
		SortBy:         "updated_at",
		SortDir:        "DESC",
		Page:           page,
		PageSize:       pageSize,
	}

	items, total, err := h.svc.ListUserFiles(c.Request().Context(), params)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"files": items,
		"total": total,
	})
}
