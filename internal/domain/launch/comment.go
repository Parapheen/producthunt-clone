package launch

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents a launch comment. Only one level of depth is supported.
// Replies can only target a top-level comment (ParentID == nil for top-level).
type Comment struct {
    ID         uuid.UUID
    LaunchID   uuid.UUID
    AuthorID   uuid.UUID
    ParentID   *uuid.UUID
    ContentHTML string
    IsPinned   bool
    Upvotes    int
    CreatedAt  time.Time
}

func NewComment(launchID, authorID uuid.UUID, contentHTML string) *Comment {
    return &Comment{
        ID:          uuid.New(),
        LaunchID:    launchID,
        AuthorID:    authorID,
        ContentHTML: contentHTML,
        IsPinned:    false,
        CreatedAt:   time.Now(),
    }
}

func (c *Comment) Validate() error {
    if c.LaunchID == uuid.Nil || c.AuthorID == uuid.Nil {
        return ErrInvalidComment
    }
    if len(c.ContentHTML) == 0 {
        return ErrInvalidComment
    }
    return nil
}


