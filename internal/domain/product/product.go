package product

import (
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID
	Name string
	URL  string

	Launches []Launch
	Members  []Member
}
