package launch

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	InvalidURLSchemeError = errors.New("invalid url scheme")
	InvalidURL            = errors.New("invalid url")
	LaunchDateInPast      = errors.New("launch date is in past")
)

func (l *Launch) Validate() error {
	u, err := url.Parse(l.URL)
	if err != nil {
		return InvalidURL
	}

	if u.Scheme == "" {
		return InvalidURLSchemeError
	}

	if strings.HasPrefix(l.URL, "http://") {
		return InvalidURLSchemeError
	}

	if !strings.HasPrefix(l.URL, "https://") {
		return InvalidURLSchemeError
	}

	now := time.Now()

	if l.LaunchDate != nil {
		if now.After(*l.LaunchDate) {
			return LaunchDateInPast
		}
	}

	return nil
}
