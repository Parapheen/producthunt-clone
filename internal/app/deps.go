package app

import "context"

type TelegramClient interface {
	Send(ctx context.Context, chatID string, message string) error
}
