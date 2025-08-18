package product

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	ProductNameMaxLength = 255
	ProductURLMaxLength  = 255
)

func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrProductNameEmpty
	}

	if p.URL == "" {
		return ErrProductURLEmpty
	}

	if len(p.Name) > ProductNameMaxLength {
		return ErrProductNameTooLong
	}

	if len(p.URL) > ProductURLMaxLength {
		return ErrProductURLTooLong
	}

	u, err := url.Parse(p.URL)
	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme == "" {
		return ErrInvalidURLScheme
	}

	if strings.HasPrefix(p.URL, "http://") {
		return ErrInvalidURLScheme
	}

	if !strings.HasPrefix(p.URL, "https://") {
		return ErrInvalidURLScheme
	}

	if len(p.Categories) == 0 {
		return ErrNoCategories
	}

	if len(p.Categories) > 3 {
		return ErrTooManyCategories
	}

	if utf8.RuneCountInString(p.Tagline) > 140 {
		return ErrProductTaglineTooLong
	}

	return nil
}
