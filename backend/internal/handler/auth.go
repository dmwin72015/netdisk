package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/middleware"
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
	userID, _ := requireUserID(c)
	sessionID := middleware.SessionID(c)
	if err := h.svc.Logout(c.Request().Context(), input.RefreshToken, userID, sessionID); err != nil {
		return err
	}
	return OK(c, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) OAuthRedirect(c echo.Context) error {
	provider := c.Param("provider")
	authURL, err := h.svc.OAuthAuthorize(c.Request().Context(), provider)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, authURL)
}

type oauthCallbackResponse struct {
	User   *service.UserResponse `json:"user"`
	Tokens *service.TokenPair    `json:"tokens"`
	New    bool                  `json:"new"`
}

func (h *AuthHandler) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return model.ErrInvalidInput
	}

	result, err := h.svc.OAuthCallback(c.Request().Context(), provider, code, state)
	if err != nil {
		// Redirect to frontend with error on failure
		frontendURL := h.svc.Cfg().OAuth2.FrontendURL
		return c.Redirect(http.StatusFound, frontendURL+"/oauth/callback?error="+url.QueryEscape(err.Error()))
	}

	frontendURL := h.svc.Cfg().OAuth2.FrontendURL
	redirectURL := fmt.Sprintf(
		"%s/oauth/callback?accessToken=%s&refreshToken=%s",
		frontendURL,
		url.QueryEscape(result.Tokens.AccessToken),
		url.QueryEscape(result.Tokens.RefreshToken),
	)
	return c.Redirect(http.StatusFound, redirectURL)
}
