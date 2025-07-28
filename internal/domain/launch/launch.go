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
	State
	Slug       string
	LaunchDate *time.Time
	Upvotes    int
	UpdatedAt  time.Time
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
	if l.State == Draft {
		l.State = Review
	} else if l.State == Review {
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
