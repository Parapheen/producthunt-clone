package launch

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LaunchRepository interface {
	GetBySlug(ctx context.Context, slug string) (*Launch, error)
	// GetByProductAndSlug retrieves a launch by product ID and its slug (unique per product)
	GetByProductAndSlug(ctx context.Context, productID uuid.UUID, slug string) (*Launch, error)
	GetByState(ctx context.Context, states []State) ([]*Launch, error)
	GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*Launch, error)
	Update(ctx context.Context, launch *Launch) error
	GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Launch, error)
	GetByProduct(ctx context.Context, productID uuid.UUID) ([]*Launch, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Launch, error)
	GetFeed(ctx context.Context, period string, limit, offset int) ([]*Launch, error)
	Create(ctx context.Context, launch *Launch) error
	Delete(ctx context.Context, launchID uuid.UUID) error

	// Media management
	AddMedia(ctx context.Context, launchID uuid.UUID, url string) error
	ReplaceMedia(ctx context.Context, launchID uuid.UUID, urls []string) error
	GetMediaByLaunchIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]string, error)

	// Avatar image
	UpdateImageURL(ctx context.Context, launchID uuid.UUID, imageURL string) error

	// Upvotes
	ToggleUpvote(ctx context.Context, launchID uuid.UUID, userID uuid.UUID) (upvoted bool, count int, err error)
	// GetUpvotedByUserForLaunchIDs returns map of launchID->true for items upvoted by the user
	GetUpvotedByUserForLaunchIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error)

	// Comments
	CreateComment(ctx context.Context, c *Comment) error
	GetCommentsTree(ctx context.Context, launchID uuid.UUID) ([]*Comment, map[uuid.UUID][]*Comment, error)
	ToggleCommentUpvote(ctx context.Context, commentID, userID uuid.UUID) (bool, int, error)
	PinComment(ctx context.Context, commentID uuid.UUID, pinned bool) error

	// Index helpers
	GetNthByProductOrderedByCreatedAt(ctx context.Context, productID uuid.UUID, index int) (*Launch, error)
	GetIndexByProductAndLaunchID(ctx context.Context, productID, launchID uuid.UUID) (int, error)

	ListAwards(ctx context.Context) ([]*Award, error)
	AssignAward(ctx context.Context, launchID uuid.UUID, awardCode string, periodDate time.Time) error
	GetAwardsByLaunchIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]*LaunchAward, error)

	// Awards utility queries
	HasAwardForPeriod(ctx context.Context, awardCode string, periodDate time.Time) (bool, error)
	GetTopLaunchInRange(ctx context.Context, start, end time.Time) (*Launch, error)
}
