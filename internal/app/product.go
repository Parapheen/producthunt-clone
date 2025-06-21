package app

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepo product.ProductRepository
}

func NewProductService(productRepo product.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, name, url string, owner uuid.UUID) (*product.Product, error) {
	p := product.NewProduct(name, url)

	p.AddMember(&product.Member{
		UserID: owner,
		Role:   product.Owner,
	})

	err := p.Validate()
	if err != nil {
		return nil, err
	}

	err = s.productRepo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ProductService) NameExists(ctx context.Context, name string) (bool, error) {
	return s.productRepo.ExistsByName(ctx, name)
}

func (s *ProductService) URLExists(ctx context.Context, u string) (bool, error) {
	return s.productRepo.ExistsByURL(ctx, u)
}
