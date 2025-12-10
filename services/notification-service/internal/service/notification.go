package service

import (
	"context"
	"errors"
	"fmt"

	"todoapp/services/notification-service/internal/domain/events"
	"todoapp/services/notification-service/internal/ports"
)

type TemplateRenderer interface {
	RenderTaskCreated(event events.TaskEvent) (TemplateResult, error)
	RenderTaskCompleted(event events.TaskEvent) (TemplateResult, error)
	RenderTaskDeleted(event events.TaskEvent) (TemplateResult, error)
}

type TemplateResult struct {
	Subject string
	Body    string
}

var (
	errUnknownType   = errors.New("notification: unsupported event type")
	errMissingTarget = errors.New("notification: user email is empty")
)

type NotificationService struct {
	mailer    ports.Mailer
	templates TemplateRenderer
}

func NewNotificationService(mailer ports.Mailer, templates TemplateRenderer) *NotificationService {
	return &NotificationService{mailer: mailer, templates: templates}
}

func (s *NotificationService) Handle(ctx context.Context, event events.TaskEvent) error {
	if event.IsZero() {
		return errors.New("notification: empty event payload")
	}
	if event.UserEmail == "" {
		return errMissingTarget
	}

	tpl, err := s.render(event)
	if errors.Is(err, errUnknownType) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.mailer.Send(ctx, ports.MailRequest{
		To:       event.UserEmail,
		Subject:  tpl.Subject,
		HTMLBody: tpl.Body,
	})
}

func (s *NotificationService) render(event events.TaskEvent) (TemplateResult, error) {
	switch event.Type {
	case events.TaskCreated:
		return s.templates.RenderTaskCreated(event)
	case events.TaskCompleted:
		return s.templates.RenderTaskCompleted(event)
	case events.TaskDeleted:
		return s.templates.RenderTaskDeleted(event)
	default:
		return TemplateResult{}, fmt.Errorf("%w: %s", errUnknownType, event.Type)
	}
}
