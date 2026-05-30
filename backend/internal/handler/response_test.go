package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/netdisk/server/internal/model"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   int
		msg    string
	}{
		// Model sentinel errors
		{name: "not found", err: model.ErrNotFound, status: 404, code: 1001, msg: "not found"},
		{name: "unauthorized", err: model.ErrUnauthorized, status: 401, code: 1002, msg: "unauthorized"},
		{name: "forbidden", err: model.ErrForbidden, status: 403, code: 1003, msg: "forbidden"},
		{name: "invalid input", err: model.ErrInvalidInput, status: 400, code: 1004, msg: "invalid input"},
		{name: "already exists", err: model.ErrAlreadyExists, status: 409, code: 2001, msg: "already exists"},
		{name: "name conflict", err: model.ErrNameConflict, status: 409, code: 2004, msg: "name conflict"},
		{name: "duplicate file", err: model.ErrDuplicateFile, status: 409, code: 2005, msg: "duplicate file"},
		{name: "same file conflict", err: model.ErrSameFileConflict, status: 409, code: 2006, msg: "same file conflict"},
		{name: "file too large", err: model.ErrFileTooLarge, status: 413, code: 2002, msg: "file exceeds size limit"},
		{name: "unsupported type", err: model.ErrUnsupportedType, status: 400, code: 2009, msg: "unsupported file type"},
		{name: "quota exceeded", err: model.ErrQuotaExceeded, status: 507, code: 2003, msg: "storage quota exceeded"},
		{name: "challenge expired", err: model.ErrChallengeExpired, status: 404, code: 3001, msg: "challenge expired"},
		{name: "challenge mismatch", err: model.ErrChallengeMismatch, status: 403, code: 3002, msg: "challenge mismatch"},
		{name: "dir not empty", err: model.ErrDirNotEmpty, status: 409, code: 2007, msg: "directory is not empty"},
		{name: "file required", err: model.ErrFileRequired, status: 400, code: 2008, msg: "file is required"},
		{name: "unsupported image", err: model.ErrUnsupportedImage, status: 400, code: 2010, msg: "only JPEG, PNG and WebP are supported"},

		// Unknown error falls through to internal
		{name: "unknown error", err: fmt.Errorf("something weird"), status: 500, code: 1005, msg: "internal error"},

		// Wrapped errors should still match via errors.Is
		{name: "wrapped not found", err: fmt.Errorf("context: %w", model.ErrNotFound), status: 404, code: 1001, msg: "not found"},
		{name: "wrapped unauthorized", err: fmt.Errorf("layer: %w", model.ErrUnauthorized), status: 401, code: 1002, msg: "unauthorized"},
		{name: "wrapped forbidden", err: fmt.Errorf("wrap: %w", model.ErrForbidden), status: 403, code: 1003, msg: "forbidden"},
		{name: "wrapped invalid input", err: fmt.Errorf("wrap: %w", model.ErrInvalidInput), status: 400, code: 1004, msg: "invalid input"},
		{name: "wrapped already exists", err: fmt.Errorf("wrap: %w", model.ErrAlreadyExists), status: 409, code: 2001, msg: "already exists"},
		{name: "wrapped name conflict", err: fmt.Errorf("wrap: %w", model.ErrNameConflict), status: 409, code: 2004, msg: "name conflict"},
		{name: "wrapped duplicate file", err: fmt.Errorf("wrap: %w", model.ErrDuplicateFile), status: 409, code: 2005, msg: "duplicate file"},
		{name: "wrapped same file conflict", err: fmt.Errorf("wrap: %w", model.ErrSameFileConflict), status: 409, code: 2006, msg: "same file conflict"},
		{name: "wrapped file too large", err: fmt.Errorf("wrap: %w", model.ErrFileTooLarge), status: 413, code: 2002, msg: "file exceeds size limit"},
		{name: "wrapped unsupported type", err: fmt.Errorf("wrap: %w", model.ErrUnsupportedType), status: 400, code: 2009, msg: "unsupported file type"},
		{name: "wrapped quota exceeded", err: fmt.Errorf("wrap: %w", model.ErrQuotaExceeded), status: 507, code: 2003, msg: "storage quota exceeded"},
		{name: "wrapped challenge expired", err: fmt.Errorf("wrap: %w", model.ErrChallengeExpired), status: 404, code: 3001, msg: "challenge expired"},
		{name: "wrapped challenge mismatch", err: fmt.Errorf("wrap: %w", model.ErrChallengeMismatch), status: 403, code: 3002, msg: "challenge mismatch"},
		{name: "wrapped dir not empty", err: fmt.Errorf("wrap: %w", model.ErrDirNotEmpty), status: 409, code: 2007, msg: "directory is not empty"},
		{name: "wrapped file required", err: fmt.Errorf("wrap: %w", model.ErrFileRequired), status: 400, code: 2008, msg: "file is required"},
		{name: "wrapped unsupported image", err: fmt.Errorf("wrap: %w", model.ErrUnsupportedImage), status: 400, code: 2010, msg: "only JPEG, PNG and WebP are supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, msg := MapError(tt.err)
			if status != tt.status || code != tt.code || msg != tt.msg {
				t.Errorf("MapError(%v) = (%d, %d, %q), want (%d, %d, %q)",
					tt.err, status, code, msg, tt.status, tt.code, tt.msg)
			}
		})
	}
}

func TestMapError_EchoHTTPError(t *testing.T) {
	tests := []struct {
		name   string
		err    *echo.HTTPError
		status int
		code   int
		msg    string
	}{
		{
			name:   "echo 400 bad request",
			err:    echo.NewHTTPError(http.StatusBadRequest, "bad request"),
			status: 400,
			code:   1004,
			msg:    "bad request",
		},
		{
			name:   "echo 401",
			err:    echo.NewHTTPError(http.StatusUnauthorized, "auth required"),
			status: 401,
			code:   1002,
			msg:    "auth required",
		},
		{
			name:   "echo 403",
			err:    echo.NewHTTPError(http.StatusForbidden, "no access"),
			status: 403,
			code:   1003,
			msg:    "no access",
		},
		{
			name:   "echo 404",
			err:    echo.NewHTTPError(http.StatusNotFound, "missing"),
			status: 404,
			code:   1001,
			msg:    "missing",
		},
		{
			name:   "echo 413",
			err:    echo.NewHTTPError(http.StatusRequestEntityTooLarge, "too big"),
			status: 413,
			code:   2002,
			msg:    "too big",
		},
		{
			name:   "echo 507",
			err:    echo.NewHTTPError(http.StatusInsufficientStorage, "disk full"),
			status: 507,
			code:   2003,
			msg:    "disk full",
		},
		{
			name:   "echo 500",
			err:    echo.NewHTTPError(http.StatusInternalServerError, "boom"),
			status: 500,
			code:   1005,
			msg:    "boom",
		},
		{
			name:   "echo 418 unknown status",
			err:    echo.NewHTTPError(http.StatusTeapot, "teapot"),
			status: 418,
			code:   1004, // non-500 unknown falls to ErrCodeInvalidInput
			msg:    "teapot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, msg := MapError(tt.err)
			if status != tt.status || code != tt.code || msg != tt.msg {
				t.Errorf("MapError(%v) = (%d, %d, %q), want (%d, %d, %q)",
					tt.err, status, code, msg, tt.status, tt.code, tt.msg)
			}
		})
	}
}

func TestMapError_NilError(t *testing.T) {
	// nil error should still go through the switch and hit default
	status, code, msg := MapError(nil)
	if status != 500 || code != 1005 || msg != "internal error" {
		t.Errorf("MapError(nil) = (%d, %d, %q), want (500, 1005, %q)",
			status, code, msg, "internal error")
	}
}
