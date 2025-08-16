package launch

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents a launch comment. Only one level of depth is supported.
// Replies can only target a top-level comment (ParentID == nil for top-level).
type Comment struct {
	ID          uuid.UUID
	LaunchID    uuid.UUID
	AuthorID    uuid.UUID
	ParentID    *uuid.UUID
	ContentHTML string
	Tag         CommentTag
	IsPinned    bool
	Upvotes     int
	CreatedAt   time.Time
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
	// Enforce tag for top-level comments
	if c.ParentID == nil {
		if c.Tag == "" {
			return ErrInvalidComment
		}
		switch c.Tag {
		case CommentTagIdea, CommentTagQuestion, CommentTagLike:
			// ok
		default:
			return ErrInvalidComment
		}
	}
	return nil
}

// CommentTag categorizes top-level feedback
type CommentTag string

const (
	CommentTagIdea     CommentTag = "idea"
	CommentTagQuestion CommentTag = "question"
	CommentTagLike     CommentTag = "like"
)
