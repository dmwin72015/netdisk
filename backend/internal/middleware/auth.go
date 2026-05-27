package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/pkg/jwtutil"
)

const ctxKeyUserID = "auth.user_id"

func JWT(jm *jwtutil.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization")
			}
			claims, err := jm.Parse(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}
			if claims.Type != jwtutil.TokenTypeAccess {
				return echo.NewHTTPError(http.StatusUnauthorized, "wrong token type")
			}
			c.Set(ctxKeyUserID, claims.UserID)
			return next(c)
		}
	}
}

func extractToken(c echo.Context) string {
	header := c.Request().Header.Get(echo.HeaderAuthorization)
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != "" {
			return parts[1]
		}
		return ""
	}
	if c.Request().Method == http.MethodGet {
		return c.QueryParam("access_token")
	}
	return ""
}

func UserID(c echo.Context) (int64, bool) {
	v, ok := c.Get(ctxKeyUserID).(int64)
	return v, ok
}
