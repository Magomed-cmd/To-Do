package ports

import (
	"context"

	"todoapp/services/notification-service/internal/domain/events"
)

type EventHandler interface {
	Handle(ctx context.Context, event events.TaskEvent) error
}

type EventConsumer interface {
	Consume(ctx context.Context, handler EventHandler) error
}
