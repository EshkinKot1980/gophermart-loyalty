package errors

import "errors"

var (
	ErrUnexpected         = errors.New("unexpected error")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
