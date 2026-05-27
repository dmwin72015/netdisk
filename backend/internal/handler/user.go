package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/middleware"
	"github.com/netdisk/server/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func requireUserID(c echo.Context) (int64, error) {
	id, ok := middleware.UserID(c)
	if !ok {
		return 0, echo.NewHTTPError(401, "unauthorized")
	}
	return id, nil
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}
	me, err := h.svc.GetMe(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return OK(c, me)
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	if err := h.svc.UpdateProfile(c.Request().Context(), userID, input.DisplayName, input.Bio); err != nil {
		return err
	}
	return OK(c, map[string]string{"message": "profile updated"})
}

func (h *UserHandler) ChangePassword(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
	}

	if err := h.svc.ChangePassword(c.Request().Context(), userID, input.OldPassword, input.NewPassword); err != nil {
		return err
	}
	return OK(c, map[string]string{"message": "password changed"})
}

func (h *UserHandler) UploadAvatar(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(400, "file is required")
	}

	f, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(400, "cannot open file")
	}
	defer f.Close()

	// Read first 512 bytes to detect content type
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return echo.NewHTTPError(400, "cannot read file")
	}

	contentType := http.DetectContentType(buf[:n])
	if contentType != "image/jpeg" && contentType != "image/png" {
		return echo.NewHTTPError(400, "only JPEG and PNG are supported")
	}

	// Seek back to start
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return echo.NewHTTPError(500, "seek error")
	}

	url, err := h.svc.UploadAvatar(c.Request().Context(), userID, f, contentType)
	if err != nil {
		return err
	}

	return OK(c, map[string]string{"avatar_url": url})
}

func (h *UserHandler) GetAvatar(c echo.Context) error {
	userSlug := c.Param("user_slug")
	if userSlug == "" {
		return echo.NewHTTPError(400, "user_slug is required")
	}

	path, err := h.svc.GetAvatarPath(c.Request().Context(), userSlug)
	if err != nil {
		return err
	}

	return c.File(path)
}

func (h *UserHandler) ListTransactions(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	txs, total, err := h.svc.ListTransactions(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return err
	}

	return OK(c, map[string]any{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
	})
}
