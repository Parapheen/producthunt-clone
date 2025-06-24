package product

import (
	"github.com/Machiel/slugify"
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID
	Name string
	URL  string
	Slug string

	Members []*Member
}

func NewProduct(name, url string) *Product {
	id := uuid.New()

	return &Product{
		ID:   id,
		Name: name,
		URL:  url,
		Slug: slugify.Slugify(name),
	}
}

func (p *Product) AddMember(member *Member) {
	p.Members = append(p.Members, member)
}

func (p *Product) IsOwner(userID uuid.UUID) bool {
	for _, member := range p.Members {
		if member.UserID == userID && member.Role == Owner {
			return true
		}
	}
	return false
}

func (p *Product) IsNil() bool {
	return p.ID == uuid.Nil
}
