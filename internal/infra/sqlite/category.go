package sqlite

import (
	"context"
	"database/sql"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*product.Category, error) {
	query := `SELECT
			c.id, c.name, c.slug
		FROM categories c
		WHERE c.slug = $1`

	query = r.db.Rebind(query)
	c := &product.Category{}
	if err := r.db.GetContext(ctx, c, query, slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, product.ErrCategoryNotFound
		}
		return nil, err
	}

	return c, nil
}

func (r *CategoryRepository) ListAll(ctx context.Context) ([]*product.Category, error) {
    query := `SELECT c.id, c.name, c.slug FROM categories c ORDER BY c.name`
    var models []struct {
        ID   int    `db:"id"`
        Name string `db:"name"`
        Slug string `db:"slug"`
    }
    if err := r.db.SelectContext(ctx, &models, r.db.Rebind(query)); err != nil {
        return nil, err
    }
    result := make([]*product.Category, 0, len(models))
    for _, m := range models {
        result = append(result, &product.Category{ID: m.ID, Name: m.Name, Slug: m.Slug})
    }
    return result, nil
}
