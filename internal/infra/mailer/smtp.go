package mailer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"path/filepath"

	"github.com/Parapheen/ph-clone/internal/pkg/config"
)

// SMTPMailer implements the Mailer interface using a real SMTP server.
type SMTPMailer struct {
	logger *slog.Logger
	cfg    config.SMTP
}

func NewSMTPMailer(logger *slog.Logger, cfg config.SMTP) *SMTPMailer {
	return &SMTPMailer{logger: logger, cfg: cfg}
}

// Send parses an HTML template, executes it with the given data, and sends it.
func (m *SMTPMailer) Send(ctx context.Context, recipient, templateFile string, data interface{}) error {
	// 1. Parse the template file.
	tmpl, err := template.New("email").
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		}).
		ParseFiles(filepath.Join("views/emails", templateFile))
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

	// 4. Execute the "htmlBody" block of the template.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return fmt.Errorf("failed to execute htmlBody template: %w", err)
	}

	// 5. Construct the full MIME message.
	// We use a multipart/alternative message to provide both HTML and plain text versions.
	fromHeader := fmt.Sprintf("From: \"justlaunch\" <%s>", m.cfg.FromAddress)
	toHeader := fmt.Sprintf("To: %s", recipient)
	subjectHeader := fmt.Sprintf("Subject: %s", subject.String())
	mimeHeader := "MIME-version: 1.0;\nContent-Type: multipart/alternative; boundary=\"boundary\"\n\n"

	msgBody := "--boundary\n"
	msgBody += "Content-Type: text/plain; charset=\"UTF-8\"\n\n"
	msgBody += plainBody.String() + "\n\n"
	msgBody += "--boundary\n"
	msgBody += "Content-Type: text/html; charset=\"UTF-8\"\n\n"
	msgBody += htmlBody.String() + "\n\n"
	msgBody += "--boundary--"

	msg := fromHeader + "\n" + toHeader + "\n" + subjectHeader + "\n" + mimeHeader + "\n" + msgBody

	// 6. Authenticate and send the email.
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	m.logger.InfoContext(ctx, "Sending real email", "to", recipient, "from", m.cfg.FromAddress, "subject", subject.String())

	// We run this in a goroutine so it doesn't block the HTTP request.
	// In a real high-volume app, this would be pushed to a background job queue.
	go func() {
		err := smtp.SendMail(addr, auth, m.cfg.FromAddress, []string{recipient}, []byte(msg))
		if err != nil {
			m.logger.Error("Failed to send email via SMTP", "to", recipient, "error", err)
		}
	}()

	return nil
}
