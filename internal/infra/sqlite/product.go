package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
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

		for _, category := range product.Categories {
			_, err := tx.ExecContext(context.WithoutCancel(ctx), `
				INSERT INTO product_categories (product_id, category_id)
				VALUES ($1, $2)
			`, product.ID, category.ID)
			if err != nil {
				return err
			}
		}

		_, err = tx.ExecContext(context.WithoutCancel(ctx), `
				INSERT INTO launches (id, product_id, name, url, state, slug)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, uuid.New(), product.ID, product.Name, product.URL, launch.Draft.String(), product.Slug)
		if err != nil {
			return err
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
	query := `
		SELECT p.id, p.name, p.url, p.slug, p.created_at
		FROM products p
		WHERE p.slug = $1`

	p := &product.Product{}
	err := r.db.GetContext(ctx, p, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}

	members, err := r.GetMembers(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Members = members

	categoryQuery := `
		SELECT c.id, c.name, c.slug
		FROM categories c
		JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = $1`

	var categories []*product.Category
	if err := r.db.SelectContext(ctx, &categories, categoryQuery, p.ID); err != nil {
		return nil, err
	}
	p.Categories = categories

	return p, nil
}

func (r *ProductRepository) GetByOwner(ctx context.Context, owner uuid.UUID) ([]*product.Product, error) {
	query := `SELECT p.id, p.name, p.url, p.slug, p.created_at, m.user_id, m.role
		FROM products p
		LEFT JOIN product_members m ON p.id = m.product_id
		WHERE m.user_id = $1
		ORDER BY p.created_at DESC`
	var products []*product.Product

	rows, err := r.db.QueryContext(ctx, query, owner)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var memberUserID uuid.UUID
		var memberRole string
		p := &product.Product{}

		err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Slug, &p.CreatedAt, &memberUserID, &memberRole)
		if err != nil {
			return nil, err
		}

		member := &product.Member{
			UserID: memberUserID,
			Role:   product.ParseRole(memberRole),
		}
		p.Members = append(p.Members, member)

		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	query := `SELECT p.id, p.name, p.url, p.slug, p.created_at, m.user_id, m.role
		FROM products p
		LEFT JOIN product_members m ON p.id = m.product_id
		WHERE p.id = $1`

	p := &product.Product{}

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var memberUserID uuid.UUID
		var memberRole string
		err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Slug, &p.CreatedAt, &memberUserID, &memberRole)
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

func (r *ProductRepository) GetMembers(ctx context.Context, productID uuid.UUID) ([]*product.Member, error) {
	query := `SELECT user_id, role
		FROM product_members m
		WHERE m.product_id = ?`

	query = r.db.Rebind(query)

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}

	members := make([]*product.Member, 0)
	for rows.Next() {
		var memberUserID uuid.UUID
		var memberRole string
		err := rows.Scan(&memberUserID, &memberRole)
		if err != nil {
			return nil, err
		}

		member := &product.Member{
			UserID: memberUserID,
			Role:   product.ParseRole(memberRole),
		}
		members = append(members, member)
	}

	return members, nil
}
