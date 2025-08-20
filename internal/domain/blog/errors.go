package blog

import "errors"

var (
	ErrInvalidTitle   = errors.New("invalid blog title")
	ErrInvalidContent = errors.New("invalid blog content")
)


