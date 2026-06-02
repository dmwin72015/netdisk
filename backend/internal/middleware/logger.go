package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func RequestLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			latency := time.Since(start)

			req := c.Request()
			res := c.Response()
			ev := logger.Info()
			if res.Status >= 500 {
				ev = logger.Error()
			} else if res.Status >= 400 {
				ev = logger.Warn()
			}
			uid, ok := UserID(c)
			ev.Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Dur("latency", latency).
				Str("remote", c.RealIP())
			if ok && uid != 0 {
				ev.Int64("user_id", uid)
			}
			if err != nil {
				ev.Err(err)
			}
			ev.Msg("http_request")
			return nil
		}
	}
}
