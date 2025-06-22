package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// LaunchModel is a model for the Launch table
type LaunchModel struct {
	ID          uuid.UUID      `db:"id"`
	ProductID   uuid.UUID      `db:"product_id"`
	Name        string         `db:"name"`
	URL         string         `db:"url"`
	Description sql.NullString `db:"description"`
	Tagline     sql.NullString `db:"tagline"`
	State       string         `db:"state"`
	Slug        string         `db:"slug"`
	LaunchDate  *time.Time     `db:"launch_date"`
}

type LaunchRepository struct {
	db *sqlx.DB
}

func NewLaunchRepository(db *sqlx.DB) *LaunchRepository {
	return &LaunchRepository{
		db: db,
	}
}

func (r *LaunchRepository) GetBySlug(ctx context.Context, slug string) (*launch.Launch, error) {
	query := `SELECT l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date
		FROM launches l
		WHERE l.slug = $1`
	l := &LaunchModel{}

	err := r.db.GetContext(ctx, l, query, slug)
	if err != nil {
		return nil, err
	}

	return &launch.Launch{
		ID:          l.ID,
		ProductID:   l.ProductID,
		Name:        l.Name,
		URL:         l.URL,
		Description: l.Description.String,
		Tagline:     l.Tagline.String,
		State:       launch.ParseState(l.State),
		Slug:        l.Slug,
		LaunchDate:  l.LaunchDate,
	}, nil
}
