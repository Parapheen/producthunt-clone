package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type TelegramClient struct {
	BotToken string
	Client   *http.Client
	logger   *slog.Logger
}

func NewTelegramClient(botToken string, logger *slog.Logger) *TelegramClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &TelegramClient{
		BotToken: botToken,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (t *TelegramClient) Send(ctx context.Context, chatID string, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal Telegram payload", slog.Any("error", err))
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to create Telegram request", slog.Any("error", err))
		return fmt.Errorf("failed to create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			t.logger.WarnContext(ctx, "Telegram request cancelled", slog.String("chatID", chatID))
			return ctx.Err()
		}
		if err, ok := err.(*url.Error); ok && err.Timeout() {
			t.logger.WarnContext(ctx, "Telegram request timed out", slog.String("chatID", chatID))
			return fmt.Errorf("telegram request timed out: %w", err)
		}
		t.logger.ErrorContext(ctx, "Failed to send message to Telegram", slog.Any("error", err), slog.String("chatID", chatID))
		return fmt.Errorf("failed to send message to telegram: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to read Telegram API response body", slog.Any("error", err), slog.String("chatID", chatID))
		return fmt.Errorf("failed to read telegram api response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.logger.ErrorContext(
			ctx,
			"Telegram API returned non-OK status",
			slog.Int("status", resp.StatusCode),
			slog.String("chatID", chatID),
			slog.String("body", string(body)),
		)
		return fmt.Errorf("telegram API error: status code %d, %s", resp.StatusCode, body)
	}

	t.logger.InfoContext(ctx, "Successfully sent Telegram notification", slog.String("chatID", chatID))
	return nil
}
