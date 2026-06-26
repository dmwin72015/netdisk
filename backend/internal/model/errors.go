package model

import "errors"

// Numeric business error codes.
const (
	// General
	ErrCodeNotFound          = 1001
	ErrCodeUnauthorized      = 1002
	ErrCodeForbidden         = 1003
	ErrCodeInvalidInput      = 1004
	ErrCodeInternal          = 1005
	ErrCodeInvalidCredentials = 1013

	// File operations
	ErrCodeAlreadyExists    = 2001
	ErrCodeFileTooLarge     = 2002
	ErrCodeQuotaExceeded    = 2003
	ErrCodeNameConflict     = 2004
	ErrCodeDuplicateFile    = 2005
	ErrCodeSameFileConflict = 2006
	ErrCodeDirNotEmpty      = 2007
	ErrCodeFileRequired     = 2008
	ErrCodeUnsupportedType  = 2009
	ErrCodeUnsupportedImage = 2010
	ErrCodeSystemFileLocked = 2011
	ErrCodeDirectoryLocked  = 2012
	ErrCodeWrongPassword    = 2013

	// Upload
	ErrCodeChallengeExpired  = 3001
	ErrCodeChallengeMismatch = 3002
)

var (
	ErrNotFound             = errors.New("not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAlreadyExists        = errors.New("already exists")
	ErrFileTooLarge         = errors.New("file exceeds size limit")
	ErrUnsupportedType      = errors.New("unsupported file type")
	ErrInternal             = errors.New("internal error")
	ErrQuotaExceeded        = errors.New("storage quota exceeded")
	ErrNameConflict         = errors.New("name conflict")
	ErrDuplicateFile        = errors.New("duplicate file")
	ErrSameFileConflict     = errors.New("same file conflict")
	ErrChallengeExpired     = errors.New("challenge expired")
	ErrChallengeMismatch    = errors.New("challenge mismatch")
	ErrDirNotEmpty          = errors.New("directory is not empty")
	ErrFileRequired         = errors.New("file is required")
	ErrUnsupportedImage     = errors.New("only JPEG, PNG and WebP are supported")
	ErrSystemFileLocked     = errors.New("system file cannot be modified")
	ErrDirectoryLocked      = errors.New("directory is locked")
	ErrWrongPassword        = errors.New("wrong password")
)

// BusinessError wraps a sentinel error so the error handler can distinguish
// business-logic failures (returned as HTTP 200 + errCode) from infrastructure
// errors (returned as HTTP 4xx/5xx).
type BusinessError struct {
	Err error
}

func (e *BusinessError) Error() string { return e.Err.Error() }
func (e *BusinessError) Unwrap() error { return e.Err }
