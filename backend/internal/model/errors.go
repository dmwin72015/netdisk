package model

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidInput       = errors.New("invalid input")
	ErrAlreadyExists      = errors.New("already exists")
	ErrFileTooLarge       = errors.New("file exceeds size limit")
	ErrUnsupportedType    = errors.New("unsupported file type")
	ErrInternal           = errors.New("internal error")
	ErrQuotaExceeded      = errors.New("storage quota exceeded")
	ErrNameConflict       = errors.New("name conflict")
	ErrDuplicateFile      = errors.New("duplicate file")
	ErrSameFileConflict   = errors.New("same file conflict")
	ErrChallengeExpired   = errors.New("challenge expired")
	ErrChallengeMismatch  = errors.New("challenge mismatch")
)
