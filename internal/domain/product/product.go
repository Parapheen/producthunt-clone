package product

import (
	"time"

	"github.com/Machiel/slugify"
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID
	Name string
	URL  string
	Slug string

	CreatedAt time.Time

	Members    []*Member
	Categories []*Category
}

func NewProduct(name, url string, categories []*Category, ownerID uuid.UUID) *Product {
	product := &Product{
		ID:         uuid.New(),
		Name:       name,
		URL:        url,
		Slug:       slugify.Slugify(name),
		CreatedAt:  time.Now(),
		Categories: categories,
		Members:    []*Member{},
	}
	product.AddMember(&Member{
		UserID: ownerID,
		Role:   Owner,
	})
	return product
}

func (p *Product) AddMember(member *Member) {
	p.Members = append(p.Members, member)
}

func (p *Product) AddCategory(category *Category) {
	if len(p.Categories) == 3 {
		return
	}
	p.Categories = append(p.Categories, category)
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
