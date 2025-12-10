package ports

import (
	"context"
	"time"

	"todoapp/pkg/events"
	analyticsv1 "todoapp/pkg/proto/analytics/v1"
)

type AnalyticsEvent struct {
	Type       analyticsv1.TaskEventType
	UserID     int64
	TaskID     int64
	Status     string
	Priority   string
	OccurredAt time.Time
}

type AnalyticsTracker interface {
	TrackTaskEvent(ctx context.Context, event AnalyticsEvent) error
}

type TaskEventPublisher interface {
	Publish(ctx context.Context, event events.TaskEvent) error
}
