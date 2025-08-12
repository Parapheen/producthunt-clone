package app

import (
	"context"
	"io"
)

// Storage abstracts binary object storage for user avatars, product images, and launch media.
// Implementations live in infra (e.g., local filesystem, S3).
type Storage interface {
    // Save stores the content under a logical path and returns a public URL.
    // The pathPrefix should be a safe, pre-sanitized logical namespace like
    // "users/<userID>/avatars" or "products/<productID>".
    Save(ctx context.Context, pathPrefix string, originalFilename string, r io.Reader) (publicURL string, err error)

    // Delete removes the object addressable by the provided public URL, if supported.
    // Implementations may choose best effort semantics.
    Delete(ctx context.Context, publicURL string) error
}


