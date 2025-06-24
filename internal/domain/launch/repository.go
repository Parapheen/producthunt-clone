package launch

import (
	"context"

	"github.com/google/uuid"
)

type LaunchRepository interface {
	GetBySlug(ctx context.Context, slug string) (*Launch, error)
	GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*Launch, error)
	Update(ctx context.Context, launch *Launch) error
}
