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
		status, code, msg := MapError(err)
		if status >= 500 {
			logger.Error().Err(err).Int("status", status).Str("path", c.Path()).Msg("handler error")
		}
		_ = c.JSON(status, map[string]any{"error": msg, "errCode": code})
	}
}

func MapError(err error) (int, int, string) {
	var he *echo.HTTPError
	if errors.As(err, &he) {
		msg, _ := he.Message.(string)
		code := mapHTTPErrorCode(he.Code)
		return he.Code, code, msg
	}
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, model.ErrCodeNotFound, "not found"
	case errors.Is(err, model.ErrUnauthorized):
		return http.StatusUnauthorized, model.ErrCodeUnauthorized, "unauthorized"
	case errors.Is(err, model.ErrForbidden):
		return http.StatusForbidden, model.ErrCodeForbidden, "forbidden"
	case errors.Is(err, model.ErrInvalidInput):
		return http.StatusBadRequest, model.ErrCodeInvalidInput, "invalid input"
	case errors.Is(err, model.ErrAlreadyExists):
		return http.StatusConflict, model.ErrCodeAlreadyExists, "already exists"
	case errors.Is(err, model.ErrNameConflict):
		return http.StatusConflict, model.ErrCodeNameConflict, "name conflict"
	case errors.Is(err, model.ErrDuplicateFile):
		return http.StatusConflict, model.ErrCodeDuplicateFile, "duplicate file"
	case errors.Is(err, model.ErrSameFileConflict):
		return http.StatusConflict, model.ErrCodeSameFileConflict, "same file conflict"
	case errors.Is(err, model.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge, model.ErrCodeFileTooLarge, "file exceeds size limit"
	case errors.Is(err, model.ErrUnsupportedType):
		return http.StatusBadRequest, model.ErrCodeUnsupportedType, "unsupported file type"
	case errors.Is(err, model.ErrQuotaExceeded):
		return http.StatusInsufficientStorage, model.ErrCodeQuotaExceeded, "storage quota exceeded"
	case errors.Is(err, model.ErrChallengeExpired):
		return http.StatusNotFound, model.ErrCodeChallengeExpired, "challenge expired"
	case errors.Is(err, model.ErrChallengeMismatch):
		return http.StatusForbidden, model.ErrCodeChallengeMismatch, "challenge mismatch"
	case errors.Is(err, model.ErrDirNotEmpty):
		return http.StatusConflict, model.ErrCodeDirNotEmpty, "directory is not empty"
	case errors.Is(err, model.ErrFileRequired):
		return http.StatusBadRequest, model.ErrCodeFileRequired, "file is required"
	case errors.Is(err, model.ErrUnsupportedImage):
		return http.StatusBadRequest, model.ErrCodeUnsupportedImage, "only JPEG, PNG and WebP are supported"
	case errors.Is(err, model.ErrSystemFileLocked):
		return http.StatusForbidden, model.ErrCodeSystemFileLocked, "system file cannot be modified"
	default:
		return http.StatusInternalServerError, model.ErrCodeInternal, "internal error"
	}
}

func mapHTTPErrorCode(status int) int {
	switch status {
	case http.StatusBadRequest:
		return model.ErrCodeInvalidInput
	case http.StatusUnauthorized:
		return model.ErrCodeUnauthorized
	case http.StatusForbidden:
		return model.ErrCodeForbidden
	case http.StatusNotFound:
		return model.ErrCodeNotFound
	case http.StatusRequestEntityTooLarge:
		return model.ErrCodeFileTooLarge
	case http.StatusInsufficientStorage:
		return model.ErrCodeQuotaExceeded
	default:
		if status >= 500 {
			return model.ErrCodeInternal
		}
		return model.ErrCodeInvalidInput
	}
}
