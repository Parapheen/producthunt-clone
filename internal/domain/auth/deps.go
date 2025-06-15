package auth

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/account"
	"github.com/Parapheen/ph-clone/internal/domain/auth/session"
	"github.com/Parapheen/ph-clone/internal/domain/user"
)

type UserService interface {
	Create(ctx context.Context, user *user.User) error
}

type AccountService interface {
	Create(ctx context.Context, account *account.Account) error
}

type SessionService interface {
	Create(ctx context.Context, session *session.Session) error
}
