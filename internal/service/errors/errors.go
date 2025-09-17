package errors

import "errors"

var (
	ErrUnexpected                 = errors.New("unexpected error")
	ErrAuthUserAlreadyExists      = errors.New("user already exists")
	ErrAuthInvalidCredentials     = errors.New("invalid credentials")
	ErrAuthInvalidToken           = errors.New("invalid token")
	ErrAuthTokenExpired           = errors.New("token expired")
	ErrOrderUploadedByUser        = errors.New("order already uploaded by user")
	ErrOrderUploadedByAnotherUser = errors.New("order already uploaded by another user")
	ErrOrderInvalidNumber         = errors.New("invalid order number")
	ErrWithdrawInvalidSum         = errors.New("invalid withdraw sum")
	ErrInsufficientFunds          = errors.New("insufficient funds")
)
