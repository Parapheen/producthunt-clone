package product

import (
	"github.com/Machiel/slugify"
	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID
	Name string
	URL  string
	Slug string

	Launches []*launch.Launch
	Members  []Member
}

func NewProduct(name, url string) *Product {
	id := uuid.New()

	return &Product{
		ID:   id,
		Name: name,
		URL:  url,
		Slug: slugify.Slugify(name),

		Launches: []*launch.Launch{
			launch.NewLaunch(id, name, url),
		},
	}
}

func (p *Product) AddMember(member *Member) {
	p.Members = append(p.Members, *member)
}
