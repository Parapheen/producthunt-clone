package product

import "errors"

var (
	ErrInvalidURLScheme      = errors.New("invalid url scheme")
	ErrInvalidURL            = errors.New("invalid url")
	ErrProductNameTooLong    = errors.New("product name is too long")
	ErrProductURLTooLong     = errors.New("product url is too long")
	ErrProductNameEmpty      = errors.New("product name is empty")
	ErrProductURLEmpty       = errors.New("product url is empty")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrNoCategories          = errors.New("no categories")
	ErrTooManyCategories     = errors.New("too many categories")
	ErrNotFound              = errors.New("not found")
	ErrProductTaglineTooLong = errors.New("product tagline is too long")
)
