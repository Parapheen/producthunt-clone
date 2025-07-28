package handler

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
)

type AuthService interface {
	GetSocialRedirectURL(provider, state string) string
	AuthenticateWithSocial(ctx context.Context, provider, code string) (*user.User, error)
	Logout(ctx context.Context, session string) error
}

type UserService interface {
	GetBySession(ctx context.Context, session string) (*user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*user.User, error)
}

type ProductService interface {
	Create(ctx context.Context, product *product.Product) error
	NameExists(ctx context.Context, name string) (bool, error)
	URLExists(ctx context.Context, u string) (bool, error)
	GetBySlug(ctx context.Context, slug string) (*product.Product, error)
	GetByOwner(ctx context.Context, owner uuid.UUID) ([]*product.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error)
	GetMembers(ctx context.Context, productID uuid.UUID) ([]*product.Member, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*product.Category, error)
}

type LaunchService interface {
	GetBySlug(ctx context.Context, slug string) (*launch.Launch, error)
	GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error)
	Update(ctx context.Context, launch *launch.Launch) error
	GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*launch.Launch, error)
	GetByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error)
	GetPublishedByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error)
	GetByID(ctx context.Context, id uuid.UUID) (*launch.Launch, error)
	GetFeed(ctx context.Context) ([]*launch.Launch, error)
	GetByState(ctx context.Context, states []launch.State) ([]*launch.Launch, error)
	Create(ctx context.Context, launch *launch.Launch) error
	Delete(ctx context.Context, launchID uuid.UUID) error
}
