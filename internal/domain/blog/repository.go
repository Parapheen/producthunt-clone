package blog

import (
	"context"

	"github.com/google/uuid"
)

type PostRepository interface {
	Create(ctx context.Context, p *Post) error
	Update(ctx context.Context, p *Post) error
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	ListPublished(ctx context.Context, limit, offset int) ([]*Post, error)
}


