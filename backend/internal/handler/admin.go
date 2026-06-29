package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/i18n"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type AdminHandler struct {
	svc       *service.AdminService
	cfg       *config.Config
	configSvc *service.SystemConfigService
	audit     *service.AuditService
}

func NewAdminHandler(svc *service.AdminService, cfg *config.Config, configSvc *service.SystemConfigService, audit *service.AuditService) *AdminHandler {
	return &AdminHandler{svc: svc, cfg: cfg, configSvc: configSvc, audit: audit}
}

func requireAdminUserID(c echo.Context) (int64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid id parameter")
	}
	return id, nil
}

func (h *AdminHandler) DashboardStats(c echo.Context) error {
	stats, err := h.svc.DashboardStats(c.Request().Context())
	if err != nil {
		return BizError(err)
	}
	return OK(c, stats)
}

func (h *AdminHandler) ListUsers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	params := service.AdminListUsersParams{
		Limit:  limit,
		Offset: offset,
		Search: c.QueryParam("search"),
		Role:   c.QueryParam("role"),
		Sort:   c.QueryParam("sort"),
	}

	if after := c.QueryParam("created_after"); after != "" {
		if t, err := time.Parse(time.RFC3339, after); err == nil {
			params.CreatedAfter = &t
		} else if t, err := time.Parse("2006-01-02", after); err == nil {
			params.CreatedAfter = &t
		}
	}
	if before := c.QueryParam("created_before"); before != "" {
		if t, err := time.Parse(time.RFC3339, before); err == nil {
			params.CreatedBefore = &t
		} else if t, err := time.Parse("2006-01-02", before); err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			params.CreatedBefore = &endOfDay
		}
	}

	items, total, err := h.svc.ListUsers(c.Request().Context(), params)
	if err != nil {
		return BizError(err)
	}

	return OK(c, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AdminHandler) CreateUser(c echo.Context) error {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return BizError(model.ErrInvalidInput)
	}
	if input.Role == "" {
		input.Role = "user"
	}
	input.Role = strings.ToLower(input.Role)
	if input.Role != "admin" && input.Role != "user" {
		return BizError(model.ErrInvalidInput)
	}

	user, err := h.svc.CreateUser(c.Request().Context(), input.Username, input.Email, input.Password, input.Role)
	if err != nil {
		return BizError(err)
	}

	adminID, _ := requireUserID(c)
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: adminID, Action: service.ActionAdminCreateUser,
		ResourceType: "user", ResourceName: input.Username,
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		Extra: map[string]any{"targetRole": input.Role},
	})
	return OK(c, user)
}

func (h *AdminHandler) GetUser(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	user, err := h.svc.GetUser(c.Request().Context(), id)
	if err != nil {
		return BizError(model.ErrNotFound)
	}

	return OK(c, user)
}

func (h *AdminHandler) UpdateUser(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		Role *string `json:"role"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if input.Role == nil {
		return BizError(model.ErrInvalidInput)
	}
	role := strings.ToLower(*input.Role)
	if role != "admin" && role != "user" {
		return BizError(model.ErrInvalidInput)
	}

	user, err := h.svc.UpdateRole(c.Request().Context(), id, role)
	if err != nil {
		return BizError(err)
	}

	adminID, _ := requireUserID(c)
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: adminID, Action: service.ActionAdminUpdateRole,
		ResourceType: "user", ResourceName: user.Username,
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		Extra: map[string]any{"targetUserID": id, "newRole": role},
	})
	return OK(c, user)
}

func (h *AdminHandler) UpdateStorageBase(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		BaseBytes int64 `json:"baseBytes"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}

	user, err := h.svc.UpdateStorageBase(c.Request().Context(), id, input.BaseBytes)
	if err != nil {
		return BizError(err)
	}

	return OK(c, user)
}

func (h *AdminHandler) DeleteUser(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.DeleteUser(c.Request().Context(), id); err != nil {
		return BizError(err)
	}

	adminID, _ := requireUserID(c)
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: adminID, Action: service.ActionAdminDeleteUser,
		ResourceType: "user",
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		Extra: map[string]any{"targetUserID": id},
	})
	return c.NoContent(204)
}

func (h *AdminHandler) ListFiles(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	params := service.AdminListFilesParams{
		Limit:    limit,
		Offset:   offset,
		Search:   c.QueryParam("search"),
		Category: c.QueryParam("category"),
		Trashed:  c.QueryParam("trashed"),
		Sort:     c.QueryParam("sort"),
	}

	items, total, err := h.svc.ListFiles(c.Request().Context(), params)
	if err != nil {
		return BizError(err)
	}

	return OK(c, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AdminHandler) StorageStats(c echo.Context) error {
	stats, err := h.svc.StorageStats(c.Request().Context())
	if err != nil {
		return BizError(err)
	}
	return OK(c, stats)
}

func (h *AdminHandler) SystemInfo(c echo.Context) error {
	defaultQuota := h.cfg.Limits.DefaultStorageQuota
	if v, ok := h.configSvc.Get("default_quota"); ok {
		switch n := v.(type) {
		case int64:
			defaultQuota = n
		case float64:
			defaultQuota = int64(n)
		}
	}

	info := map[string]any{
		"upload": map[string]any{
			"chunkSize":     h.cfg.Upload.ChunkSize,
			"maxUploadSize": h.cfg.Storage.MaxUploadSize,
		},
		"limits": map[string]any{
			"defaultStorageQuota": defaultQuota,
			"avatarMaxSize":       h.cfg.Limits.AvatarMaxSize,
		},
		"trash": map[string]any{
			"retentionDays": h.cfg.Trash.RetentionDays,
		},
		"jwt": map[string]any{
			"accessTTLMin":   h.cfg.JWT.AccessTTLMin,
			"refreshTTLHour": h.cfg.JWT.RefreshTTLHour,
		},
		"server": map[string]any{
			"port": h.cfg.Server.Port,
		},
	}
	return OK(c, info)
}

func (h *AdminHandler) DeleteFile(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.DeleteFile(c.Request().Context(), id); err != nil {
		return BizError(err)
	}

	return c.NoContent(204)
}

func (h *AdminHandler) RestoreFile(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.RestoreFile(c.Request().Context(), id); err != nil {
		return BizError(err)
	}

	return c.NoContent(204)
}

func (h *AdminHandler) ListSystemConfig(c echo.Context) error {
	items, err := h.configSvc.List(c.Request().Context())
	if err != nil {
		return BizError(err)
	}
	return OK(c, items)
}

func (h *AdminHandler) UpdateSystemConfig(c echo.Context) error {
	var input map[string]any
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if len(input) == 0 {
		return BizError(model.ErrInvalidInput)
	}
	if err := h.configSvc.SetBatch(c.Request().Context(), input); err != nil {
		return BizError(err)
	}
	items, _ := h.configSvc.List(c.Request().Context())
	return OK(c, items)
}

func (h *AdminHandler) ResetSystemConfig(c echo.Context) error {
	var input struct {
		Key string `json:"key"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if input.Key != "" {
		if err := h.configSvc.Reset(c.Request().Context(), input.Key); err != nil {
			return BizError(err)
		}
	} else {
		if err := h.configSvc.ResetAll(c.Request().Context()); err != nil {
			return BizError(err)
		}
	}
	items, _ := h.configSvc.List(c.Request().Context())
	return OK(c, items)
}

func (h *AdminHandler) CleanupQuery(c echo.Context) error {
	var input struct {
		Slug string `json:"slug"`
		Hash string `json:"hash"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if input.Slug == "" && input.Hash == "" {
		return BizError(model.ErrInvalidInput)
	}
	result, err := h.svc.CleanupQuery(c.Request().Context(), input.Slug, input.Hash)
	if err != nil {
		return BizError(err)
	}
	return OK(c, result)
}

func (h *AdminHandler) CleanupDeleteUserFile(c echo.Context) error {
	var input struct {
		UserFileID int64 `json:"userFileId"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if input.UserFileID == 0 {
		return BizError(model.ErrInvalidInput)
	}
	result, err := h.svc.CleanupDeleteUserFile(c.Request().Context(), input.UserFileID)
	if err != nil {
		return BizError(err)
	}
	return OK(c, result)
}

func (h *AdminHandler) CleanupDeletePhysicalFile(c echo.Context) error {
	var input struct {
		PhysicalFileID int64 `json:"physicalFileId"`
	}
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	if input.PhysicalFileID == 0 {
		return BizError(model.ErrInvalidInput)
	}
	result, err := h.svc.CleanupDeletePhysicalFile(c.Request().Context(), input.PhysicalFileID)
	if err != nil {
		return BizError(err)
	}
	return OK(c, result)
}

func (h *AdminHandler) ListActivityLogs(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	params := service.AdminListActivityLogsParams{
		Limit:  limit,
		Offset: offset,
		Action: c.QueryParam("action"),
		IP:     c.QueryParam("ip"),
		Locale: i18n.ResolveLocale(c),
	}

	if uid := c.QueryParam("user_id"); uid != "" {
		if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
			params.UserID = &id
		}
	}
	if from := c.QueryParam("created_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			params.CreatedFrom = &t
		} else if t, err := time.Parse("2006-01-02", from); err == nil {
			params.CreatedFrom = &t
		}
	}
	if to := c.QueryParam("created_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			params.CreatedTo = &t
		} else if t, err := time.Parse("2006-01-02", to); err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			params.CreatedTo = &endOfDay
		}
	}

	items, total, err := h.svc.ListActivityLogs(c.Request().Context(), params)
	if err != nil {
		return BizError(err)
	}

	return OK(c, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AdminHandler) ListActivityLogActions(c echo.Context) error {
	locale := i18n.ResolveLocale(c)
	items, err := h.svc.ListActivityLogActions(c.Request().Context(), locale)
	if err != nil {
		return BizError(err)
	}
	return OK(c, items)
}
