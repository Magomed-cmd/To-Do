package smtp

import (
	"context"
	"fmt"
	stdsmtp "net/smtp"
	"strings"
	"time"

	"todoapp/services/notification-service/internal/ports"
)

type Mailer struct {
	addr string
	auth stdsmtp.Auth
	from string
}

func NewMailer(host string, port int, username, password, from string) *Mailer {
	addr := fmt.Sprintf("%s:%d", host, port)

	// Only use auth if credentials are provided (Mailhog doesn't need auth)
	var auth stdsmtp.Auth
	if username != "" && password != "" {
		auth = stdsmtp.PlainAuth("", username, password, host)
	}

	if from == "" {
		from = "noreply@todoapp.local"
	}
	return &Mailer{addr: addr, auth: auth, from: from}
}

var _ ports.Mailer = (*Mailer)(nil)

func (m *Mailer) Send(ctx context.Context, req ports.MailRequest) error {
	payload := buildMessage(m.from, req)

	done := make(chan error, 1)
	go func() {
		done <- stdsmtp.SendMail(m.addr, m.auth, m.from, []string{req.To}, []byte(payload))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func buildMessage(from string, req ports.MailRequest) string {
	headers := map[string]string{
		"From":         from,
		"To":           req.To,
		"Subject":      req.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
		"Date":         time.Now().Format(time.RFC1123Z),
	}
	var builder strings.Builder
	for k, v := range headers {
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(v)
		builder.WriteString("\r\n")
	}
	builder.WriteString("\r\n")
	builder.WriteString(req.HTMLBody)
	return builder.String()
}
