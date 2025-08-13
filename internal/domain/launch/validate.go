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
	TooManyMediaFiles     = errors.New("too many media files")
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
		// Allow launch date to be today (ignore time)
		launchDate := l.LaunchDate.Truncate(24 * time.Hour)
		today := now.Truncate(24 * time.Hour)
		if launchDate.Before(today) {
			return LaunchDateInPast
		}
	}

	if len(l.Media) > 4 {
		return TooManyMediaFiles
	}

	return nil
}
