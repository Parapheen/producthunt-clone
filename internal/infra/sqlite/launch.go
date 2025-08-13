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
    ImageURL    sql.NullString `db:"image_url"`
	State       string         `db:"state"`
	Slug        string         `db:"slug"`
	LaunchDate  *time.Time     `db:"launch_date"`
	UpdatedAt   time.Time      `db:"updated_at"`
	UpvoteCount int            `db:"upvote_count"` // For aggregate queries
    CommentCount int           `db:"comment_count"`
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
        ImageURL:    l.ImageURL.String,
        Media:       []string{},
		State:       launch.ParseState(l.State),
		Slug:        l.Slug,
		LaunchDate:  l.LaunchDate,
		Upvotes:     l.UpvoteCount,
        CommentsCount: l.CommentCount,
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
        ImageURL:    sql.NullString{String: l.ImageURL, Valid: l.ImageURL != ""},
		State:       l.State.String(),
		Slug:        l.Slug,
		LaunchDate:  l.LaunchDate,
		UpdatedAt:   l.UpdatedAt,
	}
}

// Create inserts a new launch into the database.
func (r *LaunchRepository) Create(ctx context.Context, l *launch.Launch) error {
	model := fromDomain(l)
    query := `INSERT INTO launches (id, product_id, name, url, description, tagline, image_url, state, slug, launch_date)
        VALUES (:id, :product_id, :name, :url, :description, :tagline, :image_url, :state, :slug, :launch_date)`

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

// GetBySlug retrieves a launch by its slug, including its upvote count and tags.
func (r *LaunchRepository) GetBySlug(ctx context.Context, slug string) (*launch.Launch, error) {
    query := `SELECT
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date,
            COUNT(lu.launch_id) as upvote_count,
            (SELECT COUNT(*) FROM launch_comments c WHERE c.launch_id = l.id) as comment_count
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

	domainLaunch := toDomain(l)
	
	// Load media for this launch
	mediaMap, err := r.GetMediaByLaunchIDs(ctx, []uuid.UUID{domainLaunch.ID})
	if err == nil {
		domainLaunch.Media = mediaMap[domainLaunch.ID]
	}

	return domainLaunch, nil
}

// GetByProductAndSlug retrieves a launch by its product ID and slug pair
func (r *LaunchRepository) GetByProductAndSlug(ctx context.Context, productID uuid.UUID, slug string) (*launch.Launch, error) {
    query := `SELECT
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date,
            COUNT(lu.launch_id) as upvote_count,
            (SELECT COUNT(*) FROM launch_comments c WHERE c.launch_id = l.id) as comment_count
        FROM launches l
        LEFT JOIN launch_upvotes lu ON l.id = lu.launch_id
        WHERE l.product_id = ? AND l.slug = ?
        GROUP BY l.id`

    query = r.db.Rebind(query)
    l := &LaunchModel{}
    if err := r.db.GetContext(ctx, l, query, productID, slug); err != nil {
        if err == sql.ErrNoRows {
            return nil, sql.ErrNoRows
        }
        return nil, err
    }

    domainLaunch := toDomain(l)
    // Load media
    mediaMap, err := r.GetMediaByLaunchIDs(ctx, []uuid.UUID{domainLaunch.ID})
    if err == nil {
        domainLaunch.Media = mediaMap[domainLaunch.ID]
    }
    return domainLaunch, nil
}

func (r *LaunchRepository) GetByID(ctx context.Context, id uuid.UUID) (*launch.Launch, error) {
    query := `SELECT
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date,
            COUNT(lu.launch_id) as upvote_count,
            (SELECT COUNT(*) FROM launch_comments c WHERE c.launch_id = l.id) as comment_count
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

	domainLaunch := toDomain(l)
	
	// Load media for this launch
	mediaMap, err := r.GetMediaByLaunchIDs(ctx, []uuid.UUID{domainLaunch.ID})
	if err == nil {
		domainLaunch.Media = mediaMap[domainLaunch.ID]
	}

	return domainLaunch, nil
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
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date, l.updated_at,
            (SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count,
            (SELECT COUNT(*) FROM launch_comments c WHERE c.launch_id = l.id) as comment_count
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
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date, l.updated_at,
            (SELECT COUNT(*) FROM launch_upvotes lu WHERE lu.launch_id = l.id) as upvote_count,
            (SELECT COUNT(*) FROM launch_comments c WHERE c.launch_id = l.id) as comment_count
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

    // Attach media in bulk
    mediaMap, err := r.GetMediaByLaunchIDs(ctx, extractLaunchIDs(result))
    if err == nil {
        for _, dl := range result {
            dl.Media = mediaMap[dl.ID]
        }
    }

	return result, nil
}

// GetLatestByProduct retrieves the most recent launch for a given product.
func (r *LaunchRepository) GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error) {
    query := `SELECT
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date,
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

	domainLaunch := toDomain(l)
	
	// Load media for this launch
	mediaMap, err := r.GetMediaByLaunchIDs(ctx, []uuid.UUID{domainLaunch.ID})
	if err == nil {
		domainLaunch.Media = mediaMap[domainLaunch.ID]
	}

	return domainLaunch, nil
}

func (r *LaunchRepository) GetByState(ctx context.Context, states []launch.State) ([]*launch.Launch, error) {
	var query string

	if len(states) == 0 {
        query = `
			SELECT
                l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date, l.updated_at
			FROM launches l
			GROUP BY l.id
			ORDER BY l.launch_date DESC
		`
	} else {
        query = `
			SELECT
                l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date, l.updated_at
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
            l.id, l.product_id, l.name, l.url, l.description, l.tagline, l.image_url, l.state, l.slug, l.launch_date,
            COUNT(lu.launch_id) as upvote_count,
            (SELECT COUNT(*) FROM launch_comments c WHERE c.launch_id = l.id) as comment_count
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
		whereClause += " AND l.launch_date >= ? AND l.launch_date <= ?"
		startDate := now.AddDate(0, 0, -7)
		args = append(args, startDate)
		args = append(args, now)
	case "monthly":
		whereClause += " AND l.launch_date >= ? AND l.launch_date <= ?"
		startDate := now.AddDate(0, -1, 0)
		args = append(args, startDate)
		args = append(args, now)
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

    // Attach media in bulk
    mediaMap, err := r.GetMediaByLaunchIDs(ctx, extractLaunchIDs(result))
    if err == nil {
        for _, dl := range result {
            dl.Media = mediaMap[dl.ID]
        }
    }

	return result, nil
}

// Delete deletes a launch from the database.
func (r *LaunchRepository) Delete(ctx context.Context, launchID uuid.UUID) error {
	query := `DELETE FROM launches WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, launchID)
	return err
}

// --- Media ---

func (r *LaunchRepository) AddMedia(ctx context.Context, launchID uuid.UUID, url string) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO launch_media (id, launch_id, url) VALUES (?, ?, ?)`, uuid.New(), launchID, url)
    return err
}

func (r *LaunchRepository) ReplaceMedia(ctx context.Context, launchID uuid.UUID, urls []string) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Delete existing media
    if _, err := tx.ExecContext(ctx, `DELETE FROM launch_media WHERE launch_id = ?`, launchID); err != nil {
        return err
    }
    
    // Insert new media
    for _, url := range urls {
        if _, err := tx.ExecContext(ctx, `INSERT INTO launch_media (id, launch_id, url) VALUES (?, ?, ?)`, uuid.New(), launchID, url); err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

func (r *LaunchRepository) GetMediaByLaunchIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]string, error) {
    if len(ids) == 0 {
        return map[uuid.UUID][]string{}, nil
    }
    query, args, err := sqlx.In(`SELECT launch_id, url FROM launch_media WHERE launch_id IN (?) ORDER BY created_at ASC`, ids)
    if err != nil {
        return nil, err
    }
    query = r.db.Rebind(query)
    type row struct {
        LaunchID uuid.UUID `db:"launch_id"`
        URL      string    `db:"url"`
    }
    var rows []row
    if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
        return nil, err
    }
    result := make(map[uuid.UUID][]string)
    for _, r := range rows {
        result[r.LaunchID] = append(result[r.LaunchID], r.URL)
    }
    return result, nil
}

func extractLaunchIDs(launches []*launch.Launch) []uuid.UUID {
    ids := make([]uuid.UUID, 0, len(launches))
    for _, l := range launches {
        ids = append(ids, l.ID)
    }
    return ids
}

// UpdateImageURL sets the avatar image URL for a launch
func (r *LaunchRepository) UpdateImageURL(ctx context.Context, launchID uuid.UUID, imageURL string) error {
    _, err := r.db.ExecContext(ctx, `UPDATE launches SET image_url = ?, updated_at = current_timestamp WHERE id = ?`, imageURL, launchID)
    return err
}

// ToggleUpvote toggles the upvote for a user and returns whether it's upvoted now and the new count.
func (r *LaunchRepository) ToggleUpvote(ctx context.Context, launchID uuid.UUID, userID uuid.UUID) (bool, int, error) {
    upvoted := false
    var count int
    err := runInTx(ctx, r.db, func(tx *sqlx.Tx) error {
        var exists int
        if err := tx.GetContext(ctx, &exists, r.db.Rebind(`SELECT 1 FROM launch_upvotes WHERE launch_id = ? AND user_id = ?`), launchID, userID); err != nil && err != sql.ErrNoRows {
            return err
        }

        if exists == 1 {
            if _, err := tx.ExecContext(ctx, r.db.Rebind(`DELETE FROM launch_upvotes WHERE launch_id = ? AND user_id = ?`), launchID, userID); err != nil {
                return err
            }
        } else {
            if _, err := tx.ExecContext(ctx, r.db.Rebind(`INSERT INTO launch_upvotes(launch_id, user_id) VALUES(?, ?)`), launchID, userID); err != nil {
                return err
            }
            upvoted = true
        }

        if err := tx.GetContext(ctx, &count, r.db.Rebind(`SELECT COUNT(*) FROM launch_upvotes WHERE launch_id = ?`), launchID); err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return false, 0, err
    }
    return upvoted, count, nil
}

// GetUpvotedByUserForLaunchIDs returns a map of launchID->true for the IDs upvoted by the user
func (r *LaunchRepository) GetUpvotedByUserForLaunchIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
    result := make(map[uuid.UUID]bool)
    if len(ids) == 0 {
        return result, nil
    }
    query, args, err := sqlx.In(`SELECT launch_id FROM launch_upvotes WHERE user_id = ? AND launch_id IN (?)`, userID, ids)
    if err != nil {
        return nil, err
    }
    query = r.db.Rebind(query)
    var upvotedIDs []uuid.UUID
    if err := r.db.SelectContext(ctx, &upvotedIDs, query, args...); err != nil {
        return nil, err
    }
    for _, id := range upvotedIDs {
        result[id] = true
    }
    return result, nil
}

// --- Comments ---

type commentRow struct {
    ID          uuid.UUID  `db:"id"`
    LaunchID    uuid.UUID  `db:"launch_id"`
    AuthorID    uuid.UUID  `db:"author_id"`
    ParentID    *uuid.UUID `db:"parent_id"`
    ContentHTML string     `db:"content_html"`
    IsPinned    bool       `db:"is_pinned"`
    CreatedAt   time.Time  `db:"created_at"`
    Upvotes     int        `db:"upvotes"`
}

func (r *LaunchRepository) CreateComment(ctx context.Context, c *launch.Comment) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO launch_comments (id, launch_id, author_id, parent_id, content_html, is_pinned, created_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`, c.ID, c.LaunchID, c.AuthorID, c.ParentID, c.ContentHTML, c.IsPinned)
    return err
}

func (r *LaunchRepository) GetCommentsTree(ctx context.Context, launchID uuid.UUID) ([]*launch.Comment, map[uuid.UUID][]*launch.Comment, error) {
    // Fetch all comments with upvote counts for the launch
    query := r.db.Rebind(`SELECT c.id, c.launch_id, c.author_id, c.parent_id, c.content_html, c.is_pinned, c.created_at,
        COALESCE((SELECT COUNT(*) FROM launch_comment_upvotes u WHERE u.comment_id = c.id), 0) AS upvotes
        FROM launch_comments c WHERE c.launch_id = ? ORDER BY c.is_pinned DESC, c.created_at ASC`)
    var rows []commentRow
    if err := r.db.SelectContext(ctx, &rows, query, launchID); err != nil {
        return nil, nil, err
    }
    topLevel := make([]*launch.Comment, 0)
    replies := make(map[uuid.UUID][]*launch.Comment)
    for _, rrow := range rows {
        c := &launch.Comment{
            ID:          rrow.ID,
            LaunchID:    rrow.LaunchID,
            AuthorID:    rrow.AuthorID,
            ParentID:    rrow.ParentID,
            ContentHTML: rrow.ContentHTML,
            IsPinned:    rrow.IsPinned,
            Upvotes:     rrow.Upvotes,
            CreatedAt:   rrow.CreatedAt,
        }
        if rrow.ParentID == nil {
            topLevel = append(topLevel, c)
        } else {
            replies[*rrow.ParentID] = append(replies[*rrow.ParentID], c)
        }
    }
    return topLevel, replies, nil
}

func (r *LaunchRepository) ToggleCommentUpvote(ctx context.Context, commentID, userID uuid.UUID) (bool, int, error) {
    upvoted := false
    var count int
    err := runInTx(ctx, r.db, func(tx *sqlx.Tx) error {
        var exists int
        if err := tx.GetContext(ctx, &exists, r.db.Rebind(`SELECT 1 FROM launch_comment_upvotes WHERE comment_id = ? AND user_id = ?`), commentID, userID); err != nil && err != sql.ErrNoRows {
            return err
        }
        if exists == 1 {
            if _, err := tx.ExecContext(ctx, r.db.Rebind(`DELETE FROM launch_comment_upvotes WHERE comment_id = ? AND user_id = ?`), commentID, userID); err != nil {
                return err
            }
        } else {
            if _, err := tx.ExecContext(ctx, r.db.Rebind(`INSERT INTO launch_comment_upvotes(comment_id, user_id) VALUES(?, ?)`), commentID, userID); err != nil {
                return err
            }
            upvoted = true
        }
        if err := tx.GetContext(ctx, &count, r.db.Rebind(`SELECT COUNT(*) FROM launch_comment_upvotes WHERE comment_id = ?`), commentID); err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return false, 0, err
    }
    return upvoted, count, nil
}

func (r *LaunchRepository) PinComment(ctx context.Context, commentID uuid.UUID, pinned bool) error {
    _, err := r.db.ExecContext(ctx, r.db.Rebind(`UPDATE launch_comments SET is_pinned = ?, created_at = created_at WHERE id = ?`), pinned, commentID)
    return err
}
