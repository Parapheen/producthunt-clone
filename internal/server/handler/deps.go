package handler

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

type AuthService interface {
	GetSocialRedirectURL(provider, state string) string
	AuthenticateWithSocial(ctx context.Context, provider, code string) (*user.User, error)
	Logout(ctx context.Context, session string) error
}

type UserService interface {
	GetBySession(ctx context.Context, session string) (*user.User, error)
}
