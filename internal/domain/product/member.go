package product

import "github.com/google/uuid"

type Role int

const (
	Owner Role = iota
	Developer
	Designer
)

type Member struct {
	UserID uuid.UUID
	Role
}
