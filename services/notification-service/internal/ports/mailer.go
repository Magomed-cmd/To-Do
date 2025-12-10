package ports

import "context"

type MailRequest struct {
	To       string
	Subject  string
	HTMLBody string
}

type Mailer interface {
	Send(ctx context.Context, req MailRequest) error
}
