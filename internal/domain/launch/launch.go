package launch

import (
	"time"

	"github.com/Machiel/slugify"
	"github.com/google/uuid"
)

type Launch struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	Name        string
	URL         string
	Description string
	Tagline     string
	State
	Slug       string
	LaunchDate *time.Time
	Upvotes    int

	Tags []Tag
}

func NewLaunch(productID uuid.UUID, name, url string) *Launch {
	return &Launch{
		ID:        uuid.New(),
		ProductID: productID,
		Name:      name,
		URL:       url,
		State:     Draft,
		Slug:      slugify.Slugify(name),
	}
}

func (l *Launch) AddTag(tag *Tag) {
	l.Tags = append(l.Tags, *tag)
}

func (l *Launch) Publish() {
	if l.State == Draft || l.State == Review {
		l.State = Published
		now := time.Now()
		l.LaunchDate = &now
	}
}
