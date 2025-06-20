package product

import (
	"time"
)

type Launch struct {
	Name        string
	URL         string
	Description string
	Tagline     string
	State
	Slug       string
	LaunchDate *time.Time
}
