package user

import "github.com/google/uuid"

type SocialAccount struct {
	ID         uuid.UUID
	Provider   string
	ProviderID string
	Email      string
	Name       string
}
