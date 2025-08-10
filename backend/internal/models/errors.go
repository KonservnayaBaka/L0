package models

import "errors"

var (
	ErrInvalidInput  = errors.New("invalid input data")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal server error")
	ErrOrderNotFound = errors.New("order not found")
)
