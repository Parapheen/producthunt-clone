package product

import (
	"errors"

	"github.com/google/uuid"
)

type Role int

const (
	Owner Role = iota
	Developer
	Designer
)

func (r Role) String() string {
	return [...]string{"owner", "developer", "designer"}[r]
}

func ParseRole(s string) (Role, error) {
	switch s {
	case "owner":
		return Owner, nil
	case "developer":
		return Developer, nil
	case "designer":
		return Designer, nil
	default:
		return 0, errors.New("invalid role")
	}
}

type Member struct {
	UserID uuid.UUID
	Role
}
