package mailer

import (
	"context"
)

type Mailer interface {
	Send(ctx context.Context, recipient, templateFile string, data interface{}) error
}
