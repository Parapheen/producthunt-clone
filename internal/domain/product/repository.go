package product

import "context"

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
}
