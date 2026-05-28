package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInternal            = errors.New("internal error")
	ErrCacheMiss           = errors.New("cache miss")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrChecklistIncomplete = errors.New("все пункты чек-листа должны быть отмечены")
)
