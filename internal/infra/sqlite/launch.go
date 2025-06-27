package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

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
	UpvoteCount int            `db:"upvote_count"` // Added for aggregates
}

type TagModel struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Slug string `db:"slug"`
}

type LaunchRepository struct {
	db *sqlx.DB
}

func NewLaunchRepository(db *sqlx.DB) *LaunchRepository {
	return &LaunchRepository{
		db: db,
	}
}

func toDomain(l *LaunchModel, tags []launch.Tag) *launch.Launch {
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
		Upvotes:     l.UpvoteCount,
		Tags:        tags,
	}
}

func fromDomain(l *launch.Launch) *LaunchModel {
	return &LaunchModel{
		ID:          l.ID,
		ProductID:   l.ProductID,
		Name:        l.Name,
		URL:         l.URL,
		Description: sql.NullString{String: l.Description, Valid: l.Description != ""},
		Tagline:     sql.NullString{String: l.Tagline, Valid: l.Tagline != ""},
		State:       string(l.State),
		Slug:        l.Slug,
		LaunchDate:  l.LaunchDate,
	}
}

func (r *LaunchRepository) Create(ctx context.Context, l *launch.Launch) error {
	model := fromDomain(l)
	query := `INSERT INTO launches (id, product_id, name, url, description, tagline, state, slug, launch_date)
		VALUES (:id, :product_id, :name, :url, :description, :tagline, :state, :slug, :launch_date)`

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

func (r *LaunchRepository) GetBySlug(ctx context.Context, slug string) (*launch.Launch, error) {
	query := `SELECT 
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			COUNT(lu.id) as upvote_count
		FROM launches l
		LEFT JOIN launch_upvotes lu ON l.id = lu.launch_id
		WHERE l.slug = $1
		GROUP BY l.id`

	l := &LaunchModel{}
	err := r.db.GetContext(ctx, l, query, slug)
	if err != nil {
		return nil, err
	}

	tags, err := r.getTagsForLaunch(ctx, l.ID)
	if err != nil {
		return nil, err
	}

	return toDomain(l, tags), nil
}

func (r *LaunchRepository) Update(ctx context.Context, l *launch.Launch) error {
	model := fromDomain(l)
	query := `UPDATE launches
		SET name = :name, url = :url, tagline = :tagline, description = :description, state = :state, launch_date = :launch_date
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return err
}

func (r *LaunchRepository) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*launch.Launch, error) {
	query := `SELECT 
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			(SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count
		FROM launches l
		JOIN product_members pm ON l.product_id = pm.product_id
		WHERE pm.user_id = $1 AND pm.role = 'owner'
		ORDER BY l.launch_date DESC`

	var dbLaunches []*LaunchModel
	if err := r.db.SelectContext(ctx, &dbLaunches, query, ownerID); err != nil {
		return nil, err
	}

	result := make([]*launch.Launch, 0, len(dbLaunches))
	for _, l := range dbLaunches {
		// In a list view, you might not need to fetch tags for every single item
		// to avoid N+1 queries. We'll pass an empty slice for now.
		result = append(result, toDomain(l, []launch.Tag{}))
	}

	return result, nil
}

func (r *LaunchRepository) GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error) {
	query := `SELECT 
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			(SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count
		FROM launches l
		WHERE l.product_id = $1
		ORDER BY l.launch_date DESC
		LIMIT 1`

	l := &LaunchModel{}
	err := r.db.GetContext(ctx, l, query, productID)
	if err != nil {
		return nil, err
	}

	tags, err := r.getTagsForLaunch(ctx, l.ID)
	if err != nil {
		return nil, err
	}

	return toDomain(l, tags), nil
}

func (r *LaunchRepository) getTagsForLaunch(ctx context.Context, launchID uuid.UUID) ([]launch.Tag, error) {
	query := `SELECT t.id, t.name
        FROM tags t
        JOIN launch_tags lt ON t.id = lt.tag_id
        WHERE lt.launch_id = $1`

	var dbTags []*TagModel
	if err := r.db.SelectContext(ctx, &dbTags, query, launchID); err != nil {
		return nil, err
	}

	tags := make([]launch.Tag, 0, len(dbTags))
	for _, t := range dbTags {
		tags = append(tags, launch.Tag{ID: t.ID, Name: t.Name})
	}

	return tags, nil
}
