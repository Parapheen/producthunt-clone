package account

import (
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
)

type Account struct {
	ID         uuid.UUID
	Provider   string
	ProviderID string
	User       *user.User
}
