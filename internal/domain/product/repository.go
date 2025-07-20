package product

import (
	"context"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByURL(ctx context.Context, url string) (bool, error)
	GetBySlug(ctx context.Context, slug string) (*Product, error)
	GetByOwner(ctx context.Context, owner uuid.UUID) ([]*Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetMembers(ctx context.Context, productID uuid.UUID) ([]*Member, error)
}
