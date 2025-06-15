package session

import (
	"encoding/hex"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func NewSession(userID uuid.UUID) *Session {
	return &Session{
		ID:        uuid.New(),
		Token:     generateToken(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}
}

// generateToken generates a random string of length 16
func generateToken() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(rand.IntN(256))
	}
	return hex.EncodeToString(b)
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}
