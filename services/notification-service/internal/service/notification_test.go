package service

import (
	"context"
	"testing"
	"time"

	"todoapp/services/notification-service/internal/domain/events"
	"todoapp/services/notification-service/internal/ports"
)

type mailerStub struct {
	last ports.MailRequest
	err  error
}

func (m *mailerStub) Send(ctx context.Context, req ports.MailRequest) error {
	m.last = req
	return m.err
}

type templatesStub struct {
	result TemplateResult
	err    error
}

func (t templatesStub) RenderTaskCreated(event events.TaskEvent) (TemplateResult, error) {
	return t.result, t.err
}

func (t templatesStub) RenderTaskCompleted(event events.TaskEvent) (TemplateResult, error) {
	return t.result, t.err
}

func (t templatesStub) RenderTaskDeleted(event events.TaskEvent) (TemplateResult, error) {
	return t.result, t.err
}

func TestHandleSendsMail(t *testing.T) {
	mailer := &mailerStub{}
	templates := templatesStub{result: TemplateResult{Subject: "s", Body: "b"}}
	svc := NewNotificationService(mailer, templates)

	event := events.TaskEvent{
		ID:        "id",
		Type:      events.TaskCreated,
		TaskID:    1,
		UserEmail: "user@example.com",
		CreatedAt: time.Now(),
	}

	if err := svc.Handle(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mailer.last.To != "user@example.com" || mailer.last.Subject != "s" {
		t.Fatalf("mailer was not invoked with template data: %+v", mailer.last)
	}
}

func TestHandleDeletedEvent(t *testing.T) {
	mailer := &mailerStub{}
	templates := templatesStub{result: TemplateResult{Subject: "deleted", Body: "body"}}
	svc := NewNotificationService(mailer, templates)

	event := events.TaskEvent{
		ID:        "id",
		Type:      events.TaskDeleted,
		TaskID:    2,
		UserEmail: "user@example.com",
		CreatedAt: time.Now(),
	}

	if err := svc.Handle(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mailer.last.Subject != "deleted" {
		t.Fatalf("expected deleted template to be used")
	}
}

func TestHandleUnknownTypeIsIgnored(t *testing.T) {
	mailer := &mailerStub{}
	templates := templatesStub{}
	svc := NewNotificationService(mailer, templates)

	event := events.TaskEvent{
		ID:        "id",
		Type:      "task.unknown",
		TaskID:    3,
		UserEmail: "user@example.com",
		CreatedAt: time.Now(),
	}

	if err := svc.Handle(context.Background(), event); err != nil {
		t.Fatalf("expected unknown type to be ignored, got error: %v", err)
	}
	if mailer.last.To != "" {
		t.Fatalf("mailer should not be called for unknown types")
	}
}

func TestHandleMissingEmail(t *testing.T) {
	mailer := &mailerStub{}
	templates := templatesStub{}
	svc := NewNotificationService(mailer, templates)

	err := svc.Handle(context.Background(), events.TaskEvent{
		ID:        "id",
		Type:      events.TaskCreated,
		TaskID:    1,
		CreatedAt: time.Now(),
	})
	if err == nil {
		t.Fatalf("expected error for missing email")
	}
	if err != errMissingTarget {
		t.Fatalf("unexpected error: %v", err)
	}
}
