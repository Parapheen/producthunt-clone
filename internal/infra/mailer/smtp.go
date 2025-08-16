package mailer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

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
	rawSubject := strings.TrimSpace(subject.String())
	encodedSubject := mime.BEncoding.Encode("UTF-8", rawSubject)

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

	// 5. Construct the full MIME message with proper CRLF and encoded headers.
	// We use a multipart/alternative message to provide both HTML and plain text versions.
	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())

	fromHeader := fmt.Sprintf("From: \"justlaunch\" <%s>", m.cfg.FromAddress)
	toHeader := fmt.Sprintf("To: %s", recipient)
	subjectHeader := fmt.Sprintf("Subject: %s", encodedSubject)
	dateHeader := fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z))
	mimeVersionHeader := "MIME-Version: 1.0"
	contentTypeHeader := fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"", boundary)

	// Encode bodies as quoted-printable for better compatibility
	var plainQP bytes.Buffer
	qpPlain := quotedprintable.NewWriter(&plainQP)
	_, _ = qpPlain.Write([]byte(plainBody.String()))
	_ = qpPlain.Close()

	var htmlQP bytes.Buffer
	qpHTML := quotedprintable.NewWriter(&htmlQP)
	_, _ = qpHTML.Write([]byte(htmlBody.String()))
	_ = qpHTML.Close()

	var msgBuilder strings.Builder
	msgBuilder.WriteString(fromHeader)
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(toHeader)
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(subjectHeader)
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(dateHeader)
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(mimeVersionHeader)
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(contentTypeHeader)
	msgBuilder.WriteString("\r\n\r\n")

	// Plain text part
	msgBuilder.WriteString("--" + boundary + "\r\n")
	msgBuilder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	msgBuilder.WriteString(plainQP.String())
	msgBuilder.WriteString("\r\n\r\n")

	// HTML part
	msgBuilder.WriteString("--" + boundary + "\r\n")
	msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	msgBuilder.WriteString(htmlQP.String())
	msgBuilder.WriteString("\r\n")

	// Closing boundary
	msgBuilder.WriteString("--" + boundary + "--")

	msg := msgBuilder.String()

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
