package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetBySession(ctx context.Context, session string) (*User, error)
	GetByProvider(ctx context.Context, provider, providerID string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*User, error)

	CreateSession(ctx context.Context, user *User) error
	RefreshSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, session string) error

    UpdateAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) error
    UpdateBio(ctx context.Context, userID uuid.UUID, bio string) error
}
