package handler

import (
	"errors"
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

func (h *AuthHandler) OAuthUnlink(c echo.Context) error {
	provider := c.Param("provider")
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}
	if err := h.svc.OAuthUnlink(c.Request().Context(), userID, provider); err != nil {
		if errors.Is(err, model.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "cannot unlink the only login method. Please bind an email address first")
		}
		return err
	}
	return OK(c, map[string]string{"message": "unlinked"})
}

func (h *AuthHandler) OAuthRedirect(c echo.Context) error {
	provider := c.Param("provider")
	authURL, err := h.svc.OAuthAuthorize(c.Request().Context(), provider)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) OAuthBind(c echo.Context) error {
	provider := c.Param("provider")
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}
	authURL, err := h.svc.OAuthAuthorizeForBind(c.Request().Context(), provider, userID)
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
		if result != nil && result.IsBind {
			h.renderBindResult(c, result.Provider, err.Error())
			return nil
		}
		frontendURL := h.svc.Cfg().OAuth2.FrontendURL
		return c.Redirect(http.StatusFound, frontendURL+"/oauth/callback?error="+url.QueryEscape(err.Error()))
	}

	if result.IsBind {
		h.renderBindResult(c, result.Provider, "")
		return nil
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

func (h *AuthHandler) renderBindResult(c echo.Context, provider, errMsg string) {
	frontendURL := h.svc.Cfg().OAuth2.FrontendURL
	html := fmt.Sprintf(`<!DOCTYPE html><html><body><script>
		(function(){
			var msg = { bound: true, provider: %q, error: %q };
			try { if (window.opener) { window.opener.postMessage(msg, %q); } } catch(e) {}
			setTimeout(function(){ window.close(); }, 100);
		})();
	</script></body></html>`, provider, errMsg, frontendURL)
	c.HTML(http.StatusOK, html)
}
