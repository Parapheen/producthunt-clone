package launch

import (
	"context"

	"github.com/google/uuid"
)

type LaunchRepository interface {
	GetBySlug(ctx context.Context, slug string) (*Launch, error)
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
}
