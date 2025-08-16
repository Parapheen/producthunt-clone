package app

import "context"

// Mailer is the interface used by services to send emails.
type Mailer interface {
	Send(ctx context.Context, recipient, templateFile string, data interface{}) error
}

type TelegramClient interface {
	Send(ctx context.Context, chatID string, message string) error
}
