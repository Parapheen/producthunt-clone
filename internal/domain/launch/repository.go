package launch

import (
	"context"

	"github.com/google/uuid"
)

type LaunchRepository interface {
	GetBySlug(ctx context.Context, slug string) (*Launch, error)
	GetByState(ctx context.Context, states []State) ([]*Launch, error)
	GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*Launch, error)
	Update(ctx context.Context, launch *Launch) error
	GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Launch, error)
	GetByProduct(ctx context.Context, productID uuid.UUID) ([]*Launch, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Launch, error)
	GetFeed(ctx context.Context, period string, limit, offset int) ([]*Launch, error)
	Create(ctx context.Context, launch *Launch) error
	Delete(ctx context.Context, launchID uuid.UUID) error
}
