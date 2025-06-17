package user

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetBySession(ctx context.Context, session string) (*User, error)
	DeleteSession(ctx context.Context, session string) error
}
