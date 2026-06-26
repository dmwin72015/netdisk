package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/middleware"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/pkg/i18n"
)

// BizError wraps a sentinel error as a business error.
// The error handler will return HTTP 200 with the business errCode.
func BizError(err error) error {
	return &model.BusinessError{Err: err}
}

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
		lang := i18n.DetectLanguage(c.Request().Header.Get("Accept-Language"))

		// Business errors return HTTP 200 with errCode in the body.
		var be *model.BusinessError
		if errors.As(err, &be) {
			code, msg := mapBusinessError(be.Err, lang)
			_ = c.JSON(http.StatusOK, map[string]any{"code": code, "error": msg})
			return
		}

		status, code, msg := MapError(err, lang)

		ev := logger.Warn().Err(err).Int("status", status).Int("errCode", code).Str("path", c.Path()).Str("method", c.Request().Method)
		if uid, ok := middleware.UserID(c); ok && uid != 0 {
			ev.Int64("user_id", uid)
		}
		if q := c.QueryString(); q != "" {
			ev.Str("query", q)
		}
		if cl := c.Request().ContentLength; cl > 0 {
			ev.Int64("contentLength", cl)
		}
		if status >= 500 {
			ev.Str("severity", "error").Msg("handler error (5xx)")
		} else {
			ev.Str("severity", "warn").Msg("handler error (4xx)")
		}

		_ = c.JSON(status, map[string]any{"error": msg, "errCode": code})
	}
}

func MapError(err error, lang i18n.Language) (int, int, string) {
	var he *echo.HTTPError
	if errors.As(err, &he) {
		msg, _ := he.Message.(string)
		code := mapHTTPErrorCode(he.Code)
		return he.Code, code, msg
	}
	return http.StatusInternalServerError, model.ErrCodeInternal, i18n.T(i18n.MsgInternal, lang)
}

// mapBusinessError maps a business sentinel error to its errCode and i18n message.
func mapBusinessError(err error, lang i18n.Language) (int, string) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return model.ErrCodeNotFound, i18n.T(i18n.MsgNotFound, lang)
	case errors.Is(err, model.ErrUnauthorized):
		return model.ErrCodeUnauthorized, i18n.T(i18n.MsgUnauthorized, lang)
	case errors.Is(err, model.ErrForbidden):
		return model.ErrCodeForbidden, i18n.T(i18n.MsgForbidden, lang)
	case errors.Is(err, model.ErrInvalidInput):
		return model.ErrCodeInvalidInput, i18n.T(i18n.MsgInvalidInput, lang)
	case errors.Is(err, model.ErrInvalidCredentials):
		return model.ErrCodeInvalidCredentials, i18n.T(i18n.MsgInvalidCredentials, lang)
	case errors.Is(err, model.ErrAlreadyExists):
		return model.ErrCodeAlreadyExists, i18n.T(i18n.MsgAlreadyExists, lang)
	case errors.Is(err, model.ErrNameConflict):
		return model.ErrCodeNameConflict, i18n.T(i18n.MsgNameConflict, lang)
	case errors.Is(err, model.ErrDuplicateFile):
		return model.ErrCodeDuplicateFile, i18n.T(i18n.MsgDuplicateFile, lang)
	case errors.Is(err, model.ErrSameFileConflict):
		return model.ErrCodeSameFileConflict, i18n.T(i18n.MsgSameFileConflict, lang)
	case errors.Is(err, model.ErrFileTooLarge):
		return model.ErrCodeFileTooLarge, i18n.T(i18n.MsgFileTooLarge, lang)
	case errors.Is(err, model.ErrUnsupportedType):
		return model.ErrCodeUnsupportedType, i18n.T(i18n.MsgUnsupportedType, lang)
	case errors.Is(err, model.ErrQuotaExceeded):
		return model.ErrCodeQuotaExceeded, i18n.T(i18n.MsgQuotaExceeded, lang)
	case errors.Is(err, model.ErrChallengeExpired):
		return model.ErrCodeChallengeExpired, i18n.T(i18n.MsgChallengeExpired, lang)
	case errors.Is(err, model.ErrChallengeMismatch):
		return model.ErrCodeChallengeMismatch, i18n.T(i18n.MsgChallengeMismatch, lang)
	case errors.Is(err, model.ErrDirNotEmpty):
		return model.ErrCodeDirNotEmpty, i18n.T(i18n.MsgDirNotEmpty, lang)
	case errors.Is(err, model.ErrFileRequired):
		return model.ErrCodeFileRequired, i18n.T(i18n.MsgFileRequired, lang)
	case errors.Is(err, model.ErrUnsupportedImage):
		return model.ErrCodeUnsupportedImage, i18n.T(i18n.MsgUnsupportedImage, lang)
	case errors.Is(err, model.ErrSystemFileLocked):
		return model.ErrCodeSystemFileLocked, i18n.T(i18n.MsgSystemFileLocked, lang)
	case errors.Is(err, model.ErrDirectoryLocked):
		return model.ErrCodeDirectoryLocked, i18n.T(i18n.MsgDirectoryLocked, lang)
	case errors.Is(err, model.ErrWrongPassword):
		return model.ErrCodeWrongPassword, i18n.T(i18n.MsgWrongPassword, lang)
	default:
		return model.ErrCodeInternal, i18n.T(i18n.MsgInternal, lang)
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
