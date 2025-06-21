package product

import (
	"errors"
	"net/url"
	"strings"
)

const (
	ProductNameMaxLength = 255
	ProductURLMaxLength  = 255
)

var (
	InvalidURLSchemeError = errors.New("invalid url scheme")
	InvalidURL            = errors.New("invalid url")
	ProductNameTooLong    = errors.New("product name is too long")
	ProductURLTooLong     = errors.New("product url is too long")
	ProductNameEmpty      = errors.New("product name is empty")
	ProductURLEmpty       = errors.New("product url is empty")
)

func (p *Product) Validate() error {
	if p.Name == "" {
		return ProductNameEmpty
	}

	if p.URL == "" {
		return ProductURLEmpty
	}

	if len(p.Name) > ProductNameMaxLength {
		return ProductNameTooLong
	}

	if len(p.URL) > ProductURLMaxLength {
		return ProductURLTooLong
	}

	u, err := url.Parse(p.URL)
	if err != nil {
		return InvalidURL
	}

	if u.Scheme == "" {
		return InvalidURLSchemeError
	}

	if strings.HasPrefix(p.URL, "http://") {
		return InvalidURLSchemeError
	}

	if !strings.HasPrefix(p.URL, "https://") {
		return InvalidURLSchemeError
	}

	return nil
}
