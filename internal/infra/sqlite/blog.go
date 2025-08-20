package sqlite

import (
	"context"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/blog"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type blogPostModel struct {
	ID        uuid.UUID      `db:"id"`
	Title     string         `db:"title"`
	Slug      string         `db:"slug"`
	Excerpt   string         `db:"excerpt"`
	ContentMD string         `db:"content_md"`
	Published bool           `db:"published"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type BlogRepository struct {
	db *sqlx.DB
}

func NewBlogRepository(db *sqlx.DB) *BlogRepository { return &BlogRepository{db: db} }

func modelFromDomain(p *blog.Post) *blogPostModel {
	return &blogPostModel{
		ID:        p.ID,
		Title:     p.Title,
		Slug:      p.Slug,
		Excerpt:   p.Excerpt,
		ContentMD: p.ContentMD,
		Published: p.Published,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func modelToDomain(m *blogPostModel) *blog.Post {
	return &blog.Post{
		ID:        m.ID,
		Title:     m.Title,
		Slug:      m.Slug,
		Excerpt:   m.Excerpt,
		ContentMD: m.ContentMD,
		Published: m.Published,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (r *BlogRepository) Create(ctx context.Context, p *blog.Post) error {
	q := `INSERT INTO blog_posts (id, title, slug, excerpt, content_md, published, created_at, updated_at)
	VALUES (:id, :title, :slug, :excerpt, :content_md, :published, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, modelFromDomain(p))
	return err
}

func (r *BlogRepository) Update(ctx context.Context, p *blog.Post) error {
	q := `UPDATE blog_posts SET title = :title, slug = :slug, excerpt = :excerpt, content_md = :content_md, published = :published, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, q, modelFromDomain(p))
	return err
}

func (r *BlogRepository) GetBySlug(ctx context.Context, slug string) (*blog.Post, error) {
	q := `SELECT id, title, slug, excerpt, content_md, published, created_at, updated_at FROM blog_posts WHERE slug = $1 LIMIT 1`
	var m blogPostModel
	if err := r.db.GetContext(ctx, &m, q, slug); err != nil {
		return nil, err
	}
	return modelToDomain(&m), nil
}

func (r *BlogRepository) GetByID(ctx context.Context, id uuid.UUID) (*blog.Post, error) {
	q := `SELECT id, title, slug, excerpt, content_md, published, created_at, updated_at FROM blog_posts WHERE id = $1 LIMIT 1`
	var m blogPostModel
	if err := r.db.GetContext(ctx, &m, q, id); err != nil {
		return nil, err
	}
	return modelToDomain(&m), nil
}

func (r *BlogRepository) ListPublished(ctx context.Context, limit, offset int) ([]*blog.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT id, title, slug, excerpt, content_md, published, created_at, updated_at FROM blog_posts WHERE published = 1 ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	var rows []blogPostModel
	if err := r.db.SelectContext(ctx, &rows, q, limit, offset); err != nil {
		return nil, err
	}
	res := make([]*blog.Post, 0, len(rows))
	for i := range rows {
		res = append(res, modelToDomain(&rows[i]))
	}
	return res, nil
}


