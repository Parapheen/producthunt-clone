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
    ImageURL  sql.NullString `db:"image_url"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type memberModel struct {
	ProductID uuid.UUID `db:"product_id"`
	UserID    uuid.UUID `db:"user_id"`
	Role      string    `db:"role"`
}

type inviteModel struct {
    ID        uuid.UUID `db:"id"`
    ProductID uuid.UUID `db:"product_id"`
    Email     string    `db:"email"`
    Role      string    `db:"role"`
    Token     string    `db:"token"`
    Status    string    `db:"status"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
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
        productQuery := `INSERT INTO products (id, name, url, slug, tagline, image_url, created_at, updated_at)
                         VALUES (:id, :name, :url, :slug, :tagline, :image_url, :created_at, :updated_at)`
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

// GetByIDs fetches multiple products and their relations in bulk.
func (r *ProductRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*product.Product, error) {
    if len(ids) == 0 {
        return []*product.Product{}, nil
    }
    var productModels []productModel
    query, args, err := sqlx.In(`SELECT * FROM products WHERE id IN (?)`, ids)
    if err != nil {
        return nil, err
    }
    query = r.db.Rebind(query)
    if err := r.db.SelectContext(ctx, &productModels, query, args...); err != nil {
        return nil, err
    }

    // Fetch relations for all
    _, categories, err := r.fetchAllRelations(ctx, ids)
    if err != nil {
        return nil, err
    }

    result := make([]*product.Product, 0, len(productModels))
    for _, m := range productModels {
        p := toDomainProduct(&m)
        p.Categories = categories[p.ID]
        result = append(result, p)
    }
    return result, nil
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

// Invitations
func (r *ProductRepository) CreateInvitation(ctx context.Context, inv *product.Invitation) error {
    _, err := r.db.NamedExecContext(ctx, `INSERT INTO product_invitations (id, product_id, email, role, token, status) VALUES (:id, :product_id, :email, :role, :token, :status)`, map[string]interface{}{
        "id": inv.ID,
        "product_id": inv.ProductID,
        "email": inv.Email,
        "role": inv.Role.String(),
        "token": inv.Token,
        "status": string(inv.Status),
    })
    return err
}

func (r *ProductRepository) GetInvitationByToken(ctx context.Context, token string) (*product.Invitation, error) {
    var m inviteModel
    if err := r.db.GetContext(ctx, &m, `SELECT * FROM product_invitations WHERE token = $1`, token); err != nil {
        return nil, err
    }
    return &product.Invitation{
        ID:        m.ID,
        ProductID: m.ProductID,
        Email:     m.Email,
        Role:      product.ParseRole(m.Role),
        Token:     m.Token,
        Status:    product.InviteStatus(m.Status),
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }, nil
}

func (r *ProductRepository) MarkInvitationAccepted(ctx context.Context, token string) error {
    _, err := r.db.ExecContext(ctx, `UPDATE product_invitations SET status = 'accepted', updated_at = current_timestamp WHERE token = $1`, token)
    return err
}

func (r *ProductRepository) RevokeInvitation(ctx context.Context, token string) error {
    _, err := r.db.ExecContext(ctx, `UPDATE product_invitations SET status = 'revoked', updated_at = current_timestamp WHERE token = $1`, token)
    return err
}

func (r *ProductRepository) AddMember(ctx context.Context, productID, userID uuid.UUID, role product.Role) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO product_members (product_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, productID, userID, role.String())
    return err
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
        ImageURL:  model.ImageURL.String,
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
        ImageURL:  sql.NullString{String: p.ImageURL, Valid: p.ImageURL != ""},
		CreatedAt: p.CreatedAt,
		UpdatedAt: now,
	}
}

func (r *ProductRepository) UpdateImageURL(ctx context.Context, productID uuid.UUID, imageURL string) error {
    query := `UPDATE products SET image_url = $1, updated_at = current_timestamp WHERE id = $2`
    _, err := r.db.ExecContext(ctx, query, imageURL, productID)
    return err
}

func (r *ProductRepository) UpdateTagline(ctx context.Context, productID uuid.UUID, tagline string) error {
    query := `UPDATE products SET tagline = $1, updated_at = current_timestamp WHERE id = $2`
    _, err := r.db.ExecContext(ctx, query, tagline, productID)
    return err
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
