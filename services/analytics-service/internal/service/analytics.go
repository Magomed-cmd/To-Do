package service

import (
	"context"
	"fmt"
	"time"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/analytics-service/internal/domain"
	"todoapp/services/analytics-service/internal/domain/entities"
	"todoapp/services/analytics-service/internal/ports"
)

type AnalyticsService struct {
	repo ports.AnalyticsRepository
}

var _ ports.AnalyticsService = (*AnalyticsService)(nil)

func New(repo ports.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) TrackTaskEvent(ctx context.Context, input ports.TrackEventInput) error {
	if input.UserID == 0 {
		return fmt.Errorf("%w: user id is required", domain.ErrInvalidArgument)
	}

	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now()
	}

	delta, err := buildDelta(input.Type)
	if err != nil {
		return err
	}

	return s.repo.UpdateTaskMetrics(ctx, input.UserID, input.OccurredAt, delta)
}

func (s *AnalyticsService) GetDailyMetrics(ctx context.Context, req ports.DailyMetricsRequest) (*entities.DailyTaskMetrics, error) {
	if req.UserID == 0 {
		return nil, fmt.Errorf("%w: user id is required", domain.ErrInvalidArgument)
	}
	if req.Date.IsZero() {
		req.Date = time.Now()
	}
	return s.repo.GetDailyMetrics(ctx, req.UserID, req.Date)
}

func buildDelta(eventType analyticsv1.TaskEventType) (ports.MetricsDelta, error) {
	switch eventType {
	case analyticsv1.TaskEventType_TASK_EVENT_TYPE_CREATED:
		return ports.MetricsDelta{Created: 1, Total: 1}, nil
	case analyticsv1.TaskEventType_TASK_EVENT_TYPE_COMPLETED:
		return ports.MetricsDelta{Completed: 1}, nil
	case analyticsv1.TaskEventType_TASK_EVENT_TYPE_DELETED:
		return ports.MetricsDelta{Total: -1}, nil
	case analyticsv1.TaskEventType_TASK_EVENT_TYPE_UNSPECIFIED:
		fallthrough
	default:
		return ports.MetricsDelta{}, fmt.Errorf("unsupported event type: %s", eventType.String())
	}
}
