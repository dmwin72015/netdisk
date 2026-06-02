package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func requireAdminUserID(c echo.Context) (int64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, model.ErrInvalidInput
	}
	return id, nil
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

	items, total, err := h.svc.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AdminHandler) GetUser(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	user, err := h.svc.GetUser(c.Request().Context(), id)
	if err != nil {
		return model.ErrNotFound
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
		return model.ErrInvalidInput
	}
	if input.Role == nil {
		return model.ErrInvalidInput
	}

	user, err := h.svc.UpdateRole(c.Request().Context(), id, *input.Role)
	if err != nil {
		return err
	}

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
		return model.ErrInvalidInput
	}

	user, err := h.svc.UpdateStorageBase(c.Request().Context(), id, input.BaseBytes)
	if err != nil {
		return err
	}

	return OK(c, user)
}

func (h *AdminHandler) DeleteUser(c echo.Context) error {
	id, err := requireAdminUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.DeleteUser(c.Request().Context(), id); err != nil {
		return err
	}

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

	items, total, err := h.svc.ListFiles(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
