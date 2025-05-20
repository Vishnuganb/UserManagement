package errs

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrDuplicateUser = errors.New("user already exists")
	ErrInternal      = errors.New("internal error")
)
