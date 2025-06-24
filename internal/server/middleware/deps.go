package middleware

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

type UserService interface {
	GetBySession(ctx context.Context, session string) (*user.User, error)
}
