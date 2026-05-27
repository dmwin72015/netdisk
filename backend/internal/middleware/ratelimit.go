package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// isLoopback returns true for 127.0.0.0/8, ::1, and the synthetic "unknown"
// sentinel we use when RealIP() returns nothing. Loopback callers are skipped
// by the limiter so local development isn't artificially gated.
func isLoopback(ip string) bool {
	if ip == "" {
		return true
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback()
}

// RateLimit returns a per-IP fixed-window limiter backed by Redis INCR + EX.
// Window resets every `window` after the first request from a given IP. Each
// incrementing request gets a Retry-After header when over the limit.
//
// keyPrefix lets callers segregate counters across mount points (e.g. "auth"
// for stricter login limits vs. "api" for normal endpoints).
func RateLimit(rdb *redis.Client, keyPrefix string, limit int, window time.Duration) echo.MiddlewareFunc {
	if limit <= 0 || window <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if isLoopback(ip) {
				return next(c)
			}
			key := fmt.Sprintf("nd:rate:%s:%s", ip, keyPrefix)
			ctx := c.Request().Context()

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				// Fail-open on Redis hiccup — never punish the user for our infra failure.
				return next(c)
			}
			// Only set the TTL when the counter was just created; calling Expire
			// on each request would refresh the TTL indefinitely and the window
			// would never actually close.
			if count == 1 {
				_ = rdb.Expire(ctx, key, window).Err()
			}

			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			remaining := int64(limit) - count
			if remaining < 0 {
				remaining = 0
			}
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			if count > int64(limit) {
				if ttl, err := rdb.TTL(ctx, key).Result(); err == nil && ttl > 0 {
					c.Response().Header().Set("Retry-After", strconv.Itoa(int(ttl.Seconds())))
				} else {
					c.Response().Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				}
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}
