package user

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrSessionExpired = errors.New("session expired")
)
