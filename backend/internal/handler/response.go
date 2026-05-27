package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/model"
)

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, map[string]any{"data": data})
}

func Accepted(c echo.Context, data any) error {
	return c.JSON(http.StatusAccepted, map[string]any{"data": data})
}

func EchoErrorHandler(logger zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		code, msg := MapError(err)
		if code >= 500 {
			logger.Error().Err(err).Int("status", code).Str("path", c.Path()).Msg("handler error")
		}
		_ = c.JSON(code, map[string]any{"error": msg})
	}
}

func MapError(err error) (int, string) {
	var he *echo.HTTPError
	if errors.As(err, &he) {
		return he.Code, he.Message.(string)
	}
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, model.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, model.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, model.ErrInvalidInput):
		return http.StatusBadRequest, "invalid input"
	case errors.Is(err, model.ErrAlreadyExists):
		return http.StatusConflict, "already exists"
	case errors.Is(err, model.ErrNameConflict):
		return http.StatusConflict, "name conflict"
	case errors.Is(err, model.ErrDuplicateFile):
		return http.StatusConflict, "duplicate file"
	case errors.Is(err, model.ErrSameFileConflict):
		return http.StatusConflict, "same file conflict"
	case errors.Is(err, model.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge, "file too large"
	case errors.Is(err, model.ErrQuotaExceeded):
		return http.StatusInsufficientStorage, "storage quota exceeded"
	case errors.Is(err, model.ErrChallengeExpired):
		return http.StatusNotFound, "challenge expired"
	case errors.Is(err, model.ErrChallengeMismatch):
		return http.StatusForbidden, "challenge mismatch"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
