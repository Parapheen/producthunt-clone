package app

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepo  product.ProductRepository
	categoryRepo product.CategoryRepository
}

func NewProductService(
	productRepo product.ProductRepository,
	categoryRepo product.CategoryRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, p *product.Product) error {
	err := p.Validate()
	if err != nil {
		return err
	}

	err = s.productRepo.Create(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProductService) NameExists(ctx context.Context, name string) (bool, error) {
	return s.productRepo.ExistsByName(ctx, name)
}

func (s *ProductService) URLExists(ctx context.Context, u string) (bool, error) {
	return s.productRepo.ExistsByURL(ctx, u)
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*product.Product, error) {
	return s.productRepo.GetBySlug(ctx, slug)
}

func (s *ProductService) GetByOwner(ctx context.Context, owner uuid.UUID) ([]*product.Product, error) {
	return s.productRepo.GetByOwner(ctx, owner)
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) GetMembers(ctx context.Context, productID uuid.UUID) ([]*product.Member, error) {
	return s.productRepo.GetMembers(ctx, productID)
}

func (s *ProductService) GetCategoryBySlug(ctx context.Context, slug string) (*product.Category, error) {
	return s.categoryRepo.GetBySlug(ctx, slug)
}
