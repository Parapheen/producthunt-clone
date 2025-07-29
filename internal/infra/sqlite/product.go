package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type productModel struct {
	ID        uuid.UUID      `db:"id"`
	Name      string         `db:"name"`
	URL       string         `db:"url"`
	Slug      string         `db:"slug"`
	Tagline   sql.NullString `db:"tagline"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type memberModel struct {
	ProductID uuid.UUID `db:"product_id"`
	UserID    uuid.UUID `db:"user_id"`
	Role      string    `db:"role"`
}

type categoryModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Slug string `db:"slug"`
}

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *product.Product) error {
	model := toProductModel(p)

	return runInTx(ctx, r.db, func(tx *sqlx.Tx) error {
		// Insert the main product using the complete model
		productQuery := `INSERT INTO products (id, name, url, slug, tagline, created_at, updated_at)
						 VALUES (:id, :name, :url, :slug, :tagline, :created_at, :updated_at)`
		if _, err := tx.NamedExecContext(ctx, productQuery, model); err != nil {
			return fmt.Errorf("error inserting product: %w", err)
		}

		if len(p.Members) > 0 {
			memberModels := make([]memberModel, len(p.Members))
			for i, m := range p.Members {
				memberModels[i] = toMemberModel(p.ID, m)
			}
			membersQuery := `INSERT INTO product_members (product_id, user_id, role)
							 VALUES (:product_id, :user_id, :role)`
			if _, err := tx.NamedExecContext(ctx, membersQuery, memberModels); err != nil {
				return fmt.Errorf("error inserting members: %w", err)
			}
		}

		if len(p.Categories) > 0 {
			categoriesQuery := `INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2)`
			stmt, err := tx.PrepareContext(ctx, categoriesQuery)
			if err != nil {
				return err
			}
			defer stmt.Close()

			for _, cat := range p.Categories {
				if _, err := stmt.ExecContext(ctx, p.ID, cat.ID); err != nil {
					return fmt.Errorf("error inserting category association: %w", err)
				}
			}
		}

		return nil
	})
}

func (r *ProductRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT 1 FROM products WHERE name = $1 LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ProductRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	query := `SELECT 1 FROM products WHERE url = $1 LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, url).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	var p productModel
	query := `SELECT * FROM products WHERE id = $1`
	if err := r.db.GetContext(ctx, &p, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}

	domainProduct := toDomainProduct(&p)
	if err := r.fetchRelations(ctx, domainProduct); err != nil {
		return nil, fmt.Errorf("error fetching relations for product %s: %w", id, err)
	}

	return domainProduct, nil
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*product.Product, error) {
	var p productModel
	query := `SELECT * FROM products WHERE slug = $1`
	if err := r.db.GetContext(ctx, &p, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}

	domainProduct := toDomainProduct(&p)
	if err := r.fetchRelations(ctx, domainProduct); err != nil {
		return nil, fmt.Errorf("error fetching relations for product slug %s: %w", slug, err)
	}
	return domainProduct, nil
}

// GetByOwner is optimized to prevent N+1 queries.
func (r *ProductRepository) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*product.Product, error) {
	// 1. Get all product IDs for the owner
	var productIDs []uuid.UUID
	idQuery := `SELECT product_id FROM product_members WHERE user_id = $1 AND role = 'owner' ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &productIDs, idQuery, ownerID); err != nil {
		return nil, fmt.Errorf("error getting product ids by owner: %w", err)
	}
	if len(productIDs) == 0 {
		return []*product.Product{}, nil
	}

	// 2. Fetch all products at once
	var productModels []productModel
	query, args, err := sqlx.In(`SELECT * FROM products WHERE id IN (?)`, productIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	if err := r.db.SelectContext(ctx, &productModels, query, args...); err != nil {
		return nil, fmt.Errorf("error getting products by ids: %w", err)
	}

	// 3. Fetch all relations for these products at once
	allMembers, allCategories, err := r.fetchAllRelations(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	// 4. Map and assemble the final slice
	productMap := make(map[uuid.UUID]*product.Product)
	for _, model := range productModels {
		p := toDomainProduct(&model)
		p.Members = allMembers[p.ID]
		p.Categories = allCategories[p.ID]
		productMap[p.ID] = p
	}

	// Preserve the original order
	result := make([]*product.Product, len(productIDs))
	for i, id := range productIDs {
		result[i] = productMap[id]
	}

	return result, nil
}

// fetchRelations loads relations for a single product.
func (r *ProductRepository) fetchRelations(ctx context.Context, p *product.Product) error {
	// Fetch Members
	var memberModels []memberModel
	membersQuery := `SELECT product_id, user_id, role FROM product_members WHERE product_id = $1`
	if err := r.db.SelectContext(ctx, &memberModels, membersQuery, p.ID); err != nil {
		return err
	}
	p.Members = make([]*product.Member, len(memberModels))
	for i, m := range memberModels {
		p.Members[i] = toDomainMember(&m)
	}

	// Fetch Categories
	var categoryModels []categoryModel
	categoriesQuery := `
		SELECT c.id, c.name, c.slug FROM categories c
		JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = $1`
	if err := r.db.SelectContext(ctx, &categoryModels, categoriesQuery, p.ID); err != nil {
		return err
	}
	p.Categories = make([]*product.Category, len(categoryModels))
	for i, c := range categoryModels {
		p.Categories[i] = toDomainCategory(&c)
	}

	return nil
}

// fetchAllRelations loads relations for multiple products and maps them by product ID.
func (r *ProductRepository) fetchAllRelations(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID][]*product.Member, map[uuid.UUID][]*product.Category, error) {
	// Fetch all members
	var memberModels []memberModel
	memQuery, args, err := sqlx.In(`SELECT product_id, user_id, role FROM product_members WHERE product_id IN (?)`, productIDs)
	if err != nil {
		return nil, nil, err
	}
	memQuery = r.db.Rebind(memQuery)
	if err := r.db.SelectContext(ctx, &memberModels, memQuery, args...); err != nil {
		return nil, nil, err
	}

	// Fetch all categories
	var categoryRelations []struct {
		ProductID uuid.UUID `db:"product_id"`
		categoryModel
	}
	catQuery, args, err := sqlx.In(`
		SELECT pc.product_id, c.id, c.name, c.slug FROM categories c
		JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id IN (?)`, productIDs)
	if err != nil {
		return nil, nil, err
	}
	catQuery = r.db.Rebind(catQuery)
	if err := r.db.SelectContext(ctx, &categoryRelations, catQuery, args...); err != nil {
		return nil, nil, err
	}

	// Map to product IDs
	memberMap := make(map[uuid.UUID][]*product.Member)
	for _, m := range memberModels {
		memberMap[m.ProductID] = append(memberMap[m.ProductID], toDomainMember(&m))
	}

	categoryMap := make(map[uuid.UUID][]*product.Category)
	for _, c := range categoryRelations {
		categoryMap[c.ProductID] = append(categoryMap[c.ProductID], toDomainCategory(&c.categoryModel))
	}

	return memberMap, categoryMap, nil
}

// --- Mappers ---

func toDomainProduct(model *productModel) *product.Product {
	return &product.Product{
		ID:        model.ID,
		Name:      model.Name,
		URL:       model.URL,
		Slug:      model.Slug,
		Tagline:   model.Tagline.String,
		CreatedAt: model.CreatedAt,
	}
}

func toProductModel(p *product.Product) *productModel {
	now := time.Now().UTC()
	return &productModel{
		ID:        p.ID,
		Name:      p.Name,
		URL:       p.URL,
		Slug:      p.Slug,
		Tagline:   sql.NullString{String: p.Tagline, Valid: p.Tagline != ""},
		CreatedAt: p.CreatedAt,
		UpdatedAt: now,
	}
}

func toDomainMember(model *memberModel) *product.Member {
	return &product.Member{
		UserID: model.UserID,
		Role:   product.ParseRole(model.Role),
	}
}

func toMemberModel(productID uuid.UUID, m *product.Member) memberModel {
	return memberModel{
		ProductID: productID,
		UserID:    m.UserID,
		Role:      m.Role.String(),
	}
}

func toDomainCategory(model *categoryModel) *product.Category {
	return &product.Category{
		ID:   model.ID,
		Name: model.Name,
		Slug: model.Slug,
	}
}
