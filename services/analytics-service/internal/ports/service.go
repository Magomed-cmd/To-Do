package ports

import (
	"context"
	"time"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/analytics-service/internal/domain/entities"
)

type TrackEventInput struct {
	Type       analyticsv1.TaskEventType
	UserID     int64
	TaskID     int64
	Status     string
	Priority   string
	OccurredAt time.Time
}

type DailyMetricsRequest struct {
	UserID int64
	Date   time.Time
}

type AnalyticsService interface {
	TrackTaskEvent(ctx context.Context, input TrackEventInput) error
	GetDailyMetrics(ctx context.Context, req DailyMetricsRequest) (*entities.DailyTaskMetrics, error)
}
