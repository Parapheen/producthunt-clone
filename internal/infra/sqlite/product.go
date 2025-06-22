package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *product.Product) error {
	return runInTx(ctx, r.db, func(tx *sqlx.Tx) error {
		query := `INSERT INTO products (id, name, url, slug)
		VALUES ($1, $2, $3, $4);`

		_, err := tx.ExecContext(context.WithoutCancel(ctx), query, product.ID, product.Name, product.URL, product.Slug)
		if err != nil {
			return err
		}

		for _, member := range product.Members {
			_, err := tx.ExecContext(context.WithoutCancel(ctx), `
				INSERT INTO product_members (product_id, user_id, role)
				VALUES ($1, $2, $3)
			`, product.ID, member.UserID, member.Role.String())
			if err != nil {
				return err
			}
		}

		for _, launch := range product.Launches {
			_, err := tx.ExecContext(context.WithoutCancel(ctx), `
				INSERT INTO launches (id, product_id, name, url, state, slug)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, launch.ID, product.ID, launch.Name, launch.URL, launch.State.String(), launch.Slug)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProductRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE name = $1`
	var count int

	err := r.db.QueryRowContext(context.WithoutCancel(ctx), query, name).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func (r *ProductRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE url = $1`
	var count int

	err := r.db.QueryRowContext(context.WithoutCancel(ctx), query, url).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*product.Product, error) {
	query := `SELECT p.id, p.name, p.url, p.slug, m.user_id, m.role
		FROM products p
		LEFT JOIN product_members m ON p.id = m.product_id
		WHERE p.slug = $1`
	p := &product.Product{}

	rows, err := r.db.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var memberUserID uuid.UUID
		var memberRole string
		err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Slug, &memberUserID, &memberRole)
		if err != nil {
			return nil, err
		}

		member := &product.Member{
			UserID: memberUserID,
			Role:   product.ParseRole(memberRole),
		}
		p.Members = append(p.Members, member)
	}

	return p, nil
}
