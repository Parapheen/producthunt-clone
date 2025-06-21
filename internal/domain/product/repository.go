package product

import "context"

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByURL(ctx context.Context, url string) (bool, error)
}
