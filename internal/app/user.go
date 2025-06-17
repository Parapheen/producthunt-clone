package app

import (
	"context"

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
	return s.UserRepository.GetBySession(ctx, session)
}
