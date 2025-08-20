package blog

import (
	"time"

	"github.com/Machiel/slugify"
	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID
	Title     string
	Slug      string
	Excerpt   string
	ContentMD string
	Published bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPost(title, excerpt, contentMD string, published bool) *Post {
	return &Post{
		ID:        uuid.New(),
		Title:     title,
		Slug:      slugify.Slugify(title),
		Excerpt:   excerpt,
		ContentMD: contentMD,
		Published: published,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (p *Post) Validate() error {
	if p.Title == "" {
		return ErrInvalidTitle
	}
	if p.ContentMD == "" {
		return ErrInvalidContent
	}
	return nil
}

func (p *Post) Touch() {
	p.UpdatedAt = time.Now()
}

