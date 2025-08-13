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
	Tagline     string
	URL         string
	Description string
    ImageURL    string
    // Media contains URLs to story-like images for the launch
    Media       []string
	State
	Slug       string
	LaunchDate *time.Time
	Upvotes    int
    // CommentsCount stores total number of comments (including replies)
    CommentsCount int
	UpdatedAt  time.Time
}

func NewLaunch(productID uuid.UUID, name, url string) *Launch {
	return &Launch{
		ID:        uuid.New(),
		ProductID: productID,
		Name:      name,
		URL:       url,
        Media:     []string{},
		State:     Draft,
		Slug:      slugify.Slugify(name),
	}
}

func (l *Launch) Publish() {
	if l.State == Draft || l.State == Review {
		l.State = Published
		now := time.Now()
		l.LaunchDate = &now
	}
}

func (l *Launch) IsDraft() bool {
	return l.State == Draft
}

func (l *Launch) IsReview() bool {
	return l.State == Review
}

func (l *Launch) IsDeclined() bool {
	return l.State == Declined
}

func (l *Launch) IsPublished() bool {
	return l.State == Published
}

func (l *Launch) InModeration() bool {
	return l.State == Review || l.State == Declined
}

func (l *Launch) ProceedState() {
	switch l.State {
case Draft:
		l.State = Review
	case Review:
		l.State = Published
	}

	l.UpdatedAt = time.Now()
}

func (l *Launch) Decline() {
	l.State = Declined
}

func (l *Launch) LaunchingSoon() bool {
	if l.LaunchDate == nil || !l.IsPublished() {
		return false
	}

	now := time.Now()
	launchTime := *l.LaunchDate

	// Check if the launch date is in the future AND within the next 7 days.
	isUpcoming := launchTime.After(now)

	return isUpcoming
}
