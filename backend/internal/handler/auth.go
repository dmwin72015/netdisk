package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var input service.RegisterInput
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}
	user, err := h.svc.Register(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return Created(c, user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var input service.LoginInput
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}
	user, tokens, err := h.svc.Login(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return OK(c, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var input service.RefreshInput
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}
	tokens, err := h.svc.Refresh(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return OK(c, tokens)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var input service.LogoutInput
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}
	if err := h.svc.Logout(c.Request().Context(), input.RefreshToken); err != nil {
		return err
	}
	return OK(c, map[string]string{"message": "logged out"})
}
