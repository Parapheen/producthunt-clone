package app

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
)

type UserService struct {
	UserRepository user.UserRepository
	storage        Storage
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

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
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

func (s *UserService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*user.User, error) {
	return s.UserRepository.GetByIDs(ctx, ids)
}

func (s *UserService) WithStorage(storage Storage) *UserService {
	s.storage = storage
	return s
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, originalFilename string, content io.Reader) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}
	url, err := s.storage.Save(ctx, fmt.Sprintf("users/%s/avatars", userID.String()), originalFilename, content)
	if err != nil {
		return "", err
	}
	if err := s.UserRepository.UpdateAvatarURL(ctx, userID, url); err != nil {
		return "", err
	}
	return url, nil
}

func (s *UserService) UpdateBio(ctx context.Context, userID uuid.UUID, bio string) error {
	return s.UserRepository.UpdateBio(ctx, userID, bio)
}
