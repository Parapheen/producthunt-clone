package user

import (
	"encoding/hex"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	Token     string
	ExpiresAt time.Time
}

func NewSession() *Session {
	return &Session{
		ID:        uuid.New(),
		Token:     generateToken(),
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
