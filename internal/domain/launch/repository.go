package launch

import "context"

type LaunchRepository interface {
	GetBySlug(ctx context.Context, slug string) (*Launch, error)
}

