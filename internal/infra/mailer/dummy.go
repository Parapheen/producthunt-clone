package mailer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"path/filepath"
)

// DummyMailer simulates sending emails by logging them to the console.
type DummyMailer struct {
	logger *slog.Logger
}

func NewDummyMailer(logger *slog.Logger) *DummyMailer {
	return &DummyMailer{logger: logger}
}

// Send logs the email details instead of sending a real email.
func (m *DummyMailer) Send(ctx context.Context, recipient, templateFile string, data interface{}) error {
    // 1. Parse the template file (kept consistent with SMTP mailer path under views/emails).
    tmpl, err := template.New("email").ParseFiles(filepath.Join("views/emails", templateFile))
    if err != nil {
        return fmt.Errorf("failed to parse email template %s: %w", templateFile, err)
    }

	// 2. Execute the "subject" block of the template to get the email subject.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("failed to execute subject template: %w", err)
	}

	// 3. Execute the "plainBody" block of the template.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return fmt.Errorf("failed to execute plainBody template: %w", err)
	}
	m.logger.InfoContext(ctx, "DUMMY EMAIL SENT",
		slog.String("to", recipient),
		slog.String("subject", subject.String()),
		slog.String("body", plainBody.String()),
	)
	return nil
}
