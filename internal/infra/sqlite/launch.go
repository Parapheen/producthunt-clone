package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// LaunchModel represents the database schema for a launch.
type LaunchModel struct {
	ID          uuid.UUID      `db:"id"`
	ProductID   uuid.UUID      `db:"product_id"`
	ProductSlug string         `db:"product_slug"`
	Name        string         `db:"name"`
	URL         string         `db:"url"`
	Description sql.NullString `db:"description"`
	Tagline     sql.NullString `db:"tagline"`
	State       string         `db:"state"`
	Slug        string         `db:"slug"`
	LaunchDate  *time.Time     `db:"launch_date"`
	UpdatedAt   time.Time      `db:"updated_at"`
	UpvoteCount int            `db:"upvote_count"` // For aggregate queries
}

// LaunchRepository handles database operations for launches.
type LaunchRepository struct {
	db *sqlx.DB
}

// NewLaunchRepository creates a new LaunchRepository.
func NewLaunchRepository(db *sqlx.DB) *LaunchRepository {
	return &LaunchRepository{
		db: db,
	}
}

func toDomain(l *LaunchModel) *launch.Launch {
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
		UpdatedAt:   l.UpdatedAt,
	}
}

// fromDomain converts a domain.Launch object to a LaunchModel.
func fromDomain(l *launch.Launch) *LaunchModel {
	return &LaunchModel{
		ID:          l.ID,
		ProductID:   l.ProductID,
		Name:        l.Name,
		URL:         l.URL,
		Description: sql.NullString{String: l.Description, Valid: l.Description != ""},
		Tagline:     sql.NullString{String: l.Tagline, Valid: l.Tagline != ""},
		State:       l.State.String(),
		Slug:        l.Slug,
		LaunchDate:  l.LaunchDate,
		UpdatedAt:   l.UpdatedAt,
	}
}

// Create inserts a new launch into the database.
func (r *LaunchRepository) Create(ctx context.Context, l *launch.Launch) error {
	model := fromDomain(l)
	query := `INSERT INTO launches (id, product_id, name, url, description, tagline, state, slug, launch_date)
		VALUES (:id, :product_id, :name, :url, :description, :tagline, :state, :slug, :launch_date)`

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

// GetBySlug retrieves a launch by its slug, including its upvote count and tags.
func (r *LaunchRepository) GetBySlug(ctx context.Context, slug string) (*launch.Launch, error) {
	query := `SELECT
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			COUNT(lu.launch_id) as upvote_count
		FROM launches l
		LEFT JOIN launch_upvotes lu ON l.id = lu.launch_id
		WHERE l.slug = ?
		GROUP BY l.id`

	query = r.db.Rebind(query)
	l := &LaunchModel{}
	if err := r.db.GetContext(ctx, l, query, slug); err != nil {
		if err == sql.ErrNoRows {
			// Consider defining a domain-specific error, e.g., launch.ErrNotFound
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return toDomain(l), nil
}

func (r *LaunchRepository) GetByID(ctx context.Context, id uuid.UUID) (*launch.Launch, error) {
	query := `SELECT
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			COUNT(lu.launch_id) as upvote_count
		FROM launches l
		LEFT JOIN launch_upvotes lu ON l.id = lu.launch_id
		WHERE l.id = ?
		GROUP BY l.id`

	query = r.db.Rebind(query)
	l := &LaunchModel{}
	if err := r.db.GetContext(ctx, l, query, id); err != nil {
		if err == sql.ErrNoRows {
			// Consider defining a domain-specific error, e.g., launch.ErrNotFound
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return toDomain(l), nil
}

// Update modifies an existing launch in the database.
func (r *LaunchRepository) Update(ctx context.Context, l *launch.Launch) error {
	model := fromDomain(l)
	query := `UPDATE launches
		SET 
			name = :name,
			url = :url,
			tagline = :tagline,
			description = :description,
			state = :state,
			launch_date = :launch_date,
			updated_at = current_timestamp
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return sql.ErrNoRows // Or a domain error like launch.ErrNotFound
	}

	return err
}

// GetByOwner retrieves all launches for a given owner.
func (r *LaunchRepository) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*launch.Launch, error) {
	query := `SELECT
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date, l.updated_at,
			(SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count
		FROM launches l
		JOIN product_members pm ON l.product_id = pm.product_id
		WHERE pm.user_id = ? AND pm.role = 'owner'
		ORDER BY l.launch_date DESC`

	query = r.db.Rebind(query)
	var dbLaunches []*LaunchModel
	if err := r.db.SelectContext(ctx, &dbLaunches, query, ownerID); err != nil {
		return nil, err
	}

	result := make([]*launch.Launch, 0, len(dbLaunches))
	for _, l := range dbLaunches {
		// In a list view, we pass an empty slice for tags to avoid N+1 queries.
		// A more advanced implementation might fetch tags for all launches at once.
		result = append(result, toDomain(l))
	}

	return result, nil
}

// GetByProduct retrieves all launches for a given product.
func (r *LaunchRepository) GetByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error) {
	query := `SELECT
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date, l.updated_at,
			(SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count
		FROM launches l
		WHERE l.product_id = ?
		ORDER BY l.launch_date DESC`

	query = r.db.Rebind(query)
	var dbLaunches []*LaunchModel
	if err := r.db.SelectContext(ctx, &dbLaunches, query, productID); err != nil {
		return nil, err
	}

	result := make([]*launch.Launch, 0, len(dbLaunches))
	for _, l := range dbLaunches {
		result = append(result, toDomain(l))
	}

	return result, nil
}

// GetLatestByProduct retrieves the most recent launch for a given product.
func (r *LaunchRepository) GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error) {
	query := `SELECT
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			(SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count
		FROM launches l
		WHERE l.product_id = ?
		ORDER BY l.launch_date DESC
		LIMIT 1`

	query = r.db.Rebind(query)
	l := &LaunchModel{}
	if err := r.db.GetContext(ctx, l, query, productID); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows // Or a domain error
		}
		return nil, err
	}

	return toDomain(l), nil
}

func (r *LaunchRepository) GetByState(ctx context.Context, states []launch.State) ([]*launch.Launch, error) {
	var query string

	if len(states) == 0 {
		query = `
			SELECT
				l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date, l.updated_at
			FROM launches l
			GROUP BY l.id
			ORDER BY l.launch_date DESC
		`
	} else {
		query = `
			SELECT
				l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date, l.updated_at
			FROM launches l
			WHERE l.state IN (?)
			GROUP BY l.id
			ORDER BY l.launch_date DESC
		`
	}

	query = r.db.Rebind(query)
	args := []interface{}{}

	for _, state := range states {
		args = append(args, state.String())
	}

	var dbLaunches []*LaunchModel
	if err := r.db.SelectContext(ctx, &dbLaunches, query, args...); err != nil {
		return nil, err
	}

	result := make([]*launch.Launch, 0, len(dbLaunches))
	for _, l := range dbLaunches {
		result = append(result, toDomain(l))
	}

	return result, nil
}

// GetFeed retrieves a paginated and ordered feed of launches based on a time period.
// Valid periods are "daily", "weekly", "monthly", and "all_time".
func (r *LaunchRepository) GetFeed(ctx context.Context, period string, limit, offset int) ([]*launch.Launch, error) {
	baseQuery := `
		SELECT
			l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.state, l.slug, l.launch_date,
			COUNT(lu.launch_id) as upvote_count
		FROM launches l
		LEFT JOIN launch_upvotes lu ON l.id = lu.launch_id`

	endQuery := `
		GROUP BY l.id
		ORDER BY upvote_count DESC, l.launch_date DESC
		LIMIT ? OFFSET ?`

	whereClause := " WHERE l.state = 'published'"
	args := []interface{}{}
	now := time.Now()

	switch period {
	case "weekly":
		whereClause += " AND l.launch_date >= ?"
		startDate := now.AddDate(0, 0, -7)
		args = append(args, startDate)
	case "monthly":
		whereClause += " AND l.launch_date >= ?"
		startDate := now.AddDate(0, -1, 0)
		args = append(args, startDate)
	case "all_time":
		whereClause += " AND l.launch_date <= ?"
		args = append(args, now)
	default: // Default to "daily"
		whereClause += " AND date(l.launch_date) = date(?)"
		args = append(args, now)
	}

	args = append(args, limit, offset)

	// Combine query parts and rebind for the specific SQL driver.
	query := r.db.Rebind(baseQuery + whereClause + endQuery)

	var dbLaunches []*LaunchModel
	if err := r.db.SelectContext(ctx, &dbLaunches, query, args...); err != nil {
		return nil, err
	}

	if len(dbLaunches) == 0 {
		return []*launch.Launch{}, nil
	}

	// Solve N+1 problem by fetching all tags for the retrieved launches in a single query.
	launchIDs := make([]uuid.UUID, len(dbLaunches))
	for i, l := range dbLaunches {
		launchIDs[i] = l.ID
	}

	result := make([]*launch.Launch, 0, len(dbLaunches))
	for _, l := range dbLaunches {
		// Use the pre-fetched tags, defaulting to an empty slice if none exist.
		result = append(result, toDomain(l))
	}

	return result, nil
}

// Delete deletes a launch from the database.
func (r *LaunchRepository) Delete(ctx context.Context, launchID uuid.UUID) error {
	query := `DELETE FROM launches WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, launchID)
	return err
}
