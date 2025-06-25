package app

import (
	"context"
	"database/sql"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

type UserService struct {
	UserRepository user.UserRepository
}

func NewUserService(userRepository user.UserRepository) *UserService {
	return &UserService{
		UserRepository: userRepository,
	}
}

func (s *UserService) GetBySession(ctx context.Context, session string) (*user.User, error) {
	u, err := s.UserRepository.GetBySession(ctx, session)

	switch err {
	case nil:
		if u.Session.IsExpired() {
			return nil, user.ErrSessionExpired
		}
		return u, nil
	case sql.ErrNoRows:
		return nil, user.ErrUserNotFound
	default:
		return nil, err
	}
}

func (s *UserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	u, err := s.UserRepository.GetByID(ctx, id)

	switch err {
	case nil:
		return u, nil
	case sql.ErrNoRows:
		return nil, user.ErrUserNotFound
	default:
		return nil, err
	}
}
