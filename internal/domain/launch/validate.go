package launch

import (
	"net/url"
	"strings"
	"time"
)

func (l *Launch) Validate() error {
	u, err := url.Parse(l.URL)
	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme == "" {
		return ErrInvalidURL
	}

	if strings.HasPrefix(l.URL, "http://") {
		return ErrInvalidURL
	}

	if !strings.HasPrefix(l.URL, "https://") {
		return ErrInvalidURL
	}

	now := time.Now()

	if l.LaunchDate != nil {
		// Allow launch date to be today (ignore time)
		launchDate := l.LaunchDate.Truncate(24 * time.Hour)
		today := now.Truncate(24 * time.Hour)
		if launchDate.Before(today) {
			return ErrLaunchDateInPast
		}
	}

	if len(l.Media) > 4 {
		return ErrTooManyMediaFiles
	}

	return nil
}
