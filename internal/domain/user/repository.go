package user

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetBySession(ctx context.Context, session string) (*User, error)
	GetByProvider(ctx context.Context, provider, providerID string) (*User, error)

	CreateSession(ctx context.Context, user *User) error
	RefreshSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, session string) error
}
