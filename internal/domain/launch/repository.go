package launch

import "context"

type LaunchRepository interface {
	Create(ctx context.Context, launch *Launch) error
}

