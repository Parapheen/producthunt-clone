package auth

import (
	"context"
	"fmt"

	"github.com/Parapheen/ph-clone/internal/domain/auth/session"
)

func (s *Service) Signup(ctx context.Context, socialUser *YandexUser) (*session.Session, error) {
	// open transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error opening transaction: %w", err)
	}
	defer tx.Rollback()

	// create user
	// create account
	// create session
	// commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}
	return nil, nil
}
