package handler

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/auth"
)

type AuthService interface {
	GetRedirectURL(state string) string
	Authenticate(ctx context.Context, code string) (*auth.YandexUser, error)
}
