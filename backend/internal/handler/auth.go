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
	svc   *service.AuthService
	audit *service.AuditService
}

func NewAuthHandler(svc *service.AuthService, audit *service.AuditService) *AuthHandler {
	return &AuthHandler{svc: svc, audit: audit}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var input service.RegisterInput
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	user, err := h.svc.Register(c.Request().Context(), input)
	if err != nil {
		return BizError(err)
	}
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: user.ID, Action: service.ActionRegister,
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		DeviceID: input.DeviceID,
	})
	return Created(c, user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var input service.LoginInput
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	user, tokens, err := h.svc.Login(c.Request().Context(), input)
	if err != nil {
		return BizError(err)
	}
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: user.ID, Action: service.ActionLogin,
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		DeviceID: input.DeviceID,
	})
	return OK(c, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var input service.RefreshInput
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	tokens, err := h.svc.Refresh(c.Request().Context(), input)
	if err != nil {
		return BizError(err)
	}
	return OK(c, tokens)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var input service.LogoutInput
	if err := c.Bind(&input); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	userID, _ := requireUserID(c)
	sessionID := middleware.SessionID(c)
	if err := h.svc.Logout(c.Request().Context(), input.RefreshToken, userID, sessionID); err != nil {
		return BizError(err)
	}
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: userID, Action: service.ActionLogout,
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
	})
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
		return BizError(err)
	}
	return OK(c, map[string]string{"message": "unlinked"})
}

func (h *AuthHandler) OAuthRedirect(c echo.Context) error {
	provider := c.Param("provider")
	authURL, err := h.svc.OAuthAuthorize(c.Request().Context(), provider)
	if err != nil {
		return BizError(err)
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
		return BizError(err)
	}
	return c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return BizError(model.ErrInvalidInput)
	}

	result, err := h.svc.OAuthCallback(c.Request().Context(), provider, code, state)
	frontendURL := h.svc.Cfg().OAuth2.FrontendURL

	if err != nil {
		if result != nil && result.IsBind {
			params := url.Values{}
			params.Set("mode", "bind")
			params.Set("provider", result.Provider)
			params.Set("error", err.Error())
			return c.Redirect(http.StatusFound, frontendURL+"/oauth/callback?"+params.Encode())
		}
		return c.Redirect(http.StatusFound, frontendURL+"/oauth/callback?error="+url.QueryEscape(err.Error()))
	}

	if result.EmailMatchConfirm {
		params := url.Values{}
		params.Set("mode", "email-confirm")
		params.Set("confirmToken", result.ConfirmToken)
		params.Set("email", result.MatchEmail)
		params.Set("provider", result.MatchProvider)
		return c.Redirect(http.StatusFound, frontendURL+"/oauth/callback?"+params.Encode())
	}

	if result.IsBind {
		params := url.Values{}
		params.Set("mode", "bind")
		params.Set("provider", result.Provider)
		if result.NeedReplaceConfirm {
			params.Set("needReplaceConfirm", "true")
			params.Set("replaceToken", result.ReplaceToken)
			params.Set("replaceProvider", result.ReplaceProvider)
			if result.OldProviderAccountID != "" {
				params.Set("oldProviderAccountId", result.OldProviderAccountID)
			}
			if result.OldOAuthEmail != "" {
				params.Set("oldOauthEmail", result.OldOAuthEmail)
			}
		} else if result.AlreadyBound {
			params.Set("alreadyBound", "true")
		}
		return c.Redirect(http.StatusFound, frontendURL+"/oauth/callback?"+params.Encode())
	}

	redirectURL := fmt.Sprintf(
		"%s/oauth/callback?accessToken=%s&refreshToken=%s",
		frontendURL,
		url.QueryEscape(result.Tokens.AccessToken),
		url.QueryEscape(result.Tokens.RefreshToken),
	)
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: result.UserID, Action: service.ActionOAuthLogin,
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		Extra: map[string]any{"provider": provider},
	})
	return c.Redirect(http.StatusFound, redirectURL)
}

// OAuthBindConfirmReplace is invoked by the account page (not the OAuth popup)
// after the user confirms they want to swap their current binding for the
// same provider with a new third-party account. It returns JSON; the OAuth
// popup's postMessage flow is reserved for OAuthCallback.
func (h *AuthHandler) OAuthBindConfirmReplace(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return BizError(model.ErrInvalidInput)
	}

	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	provider, err := h.svc.ConfirmOAuthBindReplace(c.Request().Context(), token, userID)
	if err != nil {
		if errors.Is(err, model.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "replace token does not belong to this user")
		}
		if errors.Is(err, model.ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, "replace token expired or invalid")
		}
		return BizError(err)
	}

	return OK(c, map[string]any{
		"message":  "replaced",
		"provider": provider,
	})
}

// OAuthEmailConfirm is called after the user confirms they want to link
// their OAuth account to an existing local user matched by email and log in.
func (h *AuthHandler) OAuthEmailConfirm(c echo.Context) error {
	confirmToken := c.QueryParam("token")
	if confirmToken == "" {
		return BizError(model.ErrInvalidInput)
	}

	tokens, err := h.svc.ConfirmEmailOAuthLink(c.Request().Context(), confirmToken)
	if err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, "confirm token expired or invalid")
		}
		return BizError(err)
	}

	return OK(c, tokens)
}


