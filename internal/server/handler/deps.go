package handler

import (
	"context"

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
}

type ProductService interface {
	Create(ctx context.Context, name, url string, owner uuid.UUID) (*product.Product, error)
	NameExists(ctx context.Context, name string) (bool, error)
	URLExists(ctx context.Context, u string) (bool, error)
}
