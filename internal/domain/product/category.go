package product

type Category struct {
	ID   int
	Name string
	Slug string
}

func NewCategory(name, slug string) *Category {
	return &Category{
		Name: name,
		Slug: slug,
	}
}
