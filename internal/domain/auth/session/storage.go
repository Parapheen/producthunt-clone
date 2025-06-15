package session

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Create(ctx context.Context, session *Session) error {
	_, err := s.db.ExecContext(context.WithoutCancel(ctx), `
		INSERT INTO sessions (id, token, user_id, expires_at)
		VALUES (:id, :token, :user_id, :expires_at)
	`, session.ID, session.Token, session.UserID, session.ExpiresAt)

	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	return nil
}
