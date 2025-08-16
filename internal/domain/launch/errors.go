package launch

import "errors"

var (
	ErrInvalidComment      = errors.New("invalid comment")
	ErrLaunchAlreadyExists = errors.New("launch already exists")
	ErrProductNotFound     = errors.New("product not found")
	ErrLaunchNotFound      = errors.New("launch not found")
	ErrLaunchDateInPast    = errors.New("launch date is in the past")
	ErrTooManyMediaFiles   = errors.New("too many media files")
	ErrInvalidURL          = errors.New("invalid url")
)
