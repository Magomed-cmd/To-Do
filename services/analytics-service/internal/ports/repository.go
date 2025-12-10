package ports

import (
	"context"
	"time"

	"todoapp/services/analytics-service/internal/domain/entities"
)

type MetricsDelta struct {
	Created   int32
	Completed int32
	Total     int32
}

type AnalyticsRepository interface {
	UpdateTaskMetrics(ctx context.Context, userID int64, date time.Time, delta MetricsDelta) error
	GetDailyMetrics(ctx context.Context, userID int64, date time.Time) (*entities.DailyTaskMetrics, error)
}
