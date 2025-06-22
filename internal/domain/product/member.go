package product

import (
	"github.com/google/uuid"
)

type Role int

const (
	Owner Role = iota
	Developer
	Designer
	Other
)

func (r Role) String() string {
	return [...]string{"owner", "developer", "designer", "other"}[r]
}

func ParseRole(s string) Role {
	switch s {
	case "owner":
		return Owner
	case "developer":
		return Developer
	case "designer":
		return Designer
	default:
		return Other
	}
}

type Member struct {
	UserID uuid.UUID
	Role
}
