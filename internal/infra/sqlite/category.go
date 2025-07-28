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
