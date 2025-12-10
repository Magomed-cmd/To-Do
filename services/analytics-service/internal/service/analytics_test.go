package service

import (
	"context"
	"errors"
	"testing"
	"time"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/analytics-service/internal/domain"
	"todoapp/services/analytics-service/internal/domain/entities"
	"todoapp/services/analytics-service/internal/ports"
)

type stubRepo struct {
	lastUserID int64
	lastDate   time.Time
	lastDelta  ports.MetricsDelta
	updateErr  error

	metrics    *entities.DailyTaskMetrics
	metricsErr error
}

func (r *stubRepo) UpdateTaskMetrics(ctx context.Context, userID int64, date time.Time, delta ports.MetricsDelta) error {
	r.lastUserID = userID
	r.lastDate = date
	r.lastDelta = delta
	return r.updateErr
}

func (r *stubRepo) GetDailyMetrics(ctx context.Context, userID int64, date time.Time) (*entities.DailyTaskMetrics, error) {
	r.lastUserID = userID
	r.lastDate = date
	if r.metrics != nil {
		return r.metrics, r.metricsErr
	}
	return &entities.DailyTaskMetrics{UserID: userID, Date: date}, r.metricsErr
}

func TestTrackTaskEventValidation(t *testing.T) {
	svc := New(&stubRepo{})

	err := svc.TrackTaskEvent(context.Background(), ports.TrackEventInput{})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument error, got %v", err)
	}
}

func TestTrackTaskEventDefaultTimestampAndDelta(t *testing.T) {
	repo := &stubRepo{}
	svc := New(repo)

	err := svc.TrackTaskEvent(context.Background(), ports.TrackEventInput{
		Type:   analyticsv1.TaskEventType_TASK_EVENT_TYPE_CREATED,
		UserID: 42,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastUserID != 42 {
		t.Fatalf("unexpected user id: %d", repo.lastUserID)
	}
	if repo.lastDelta != (ports.MetricsDelta{Created: 1, Total: 1}) {
		t.Fatalf("unexpected delta: %+v", repo.lastDelta)
	}
	if repo.lastDate.IsZero() {
		t.Fatalf("expected occurredAt to be set")
	}
}

func TestTrackTaskEventInvalidType(t *testing.T) {
	svc := New(&stubRepo{})

	err := svc.TrackTaskEvent(context.Background(), ports.TrackEventInput{
		Type:   analyticsv1.TaskEventType_TASK_EVENT_TYPE_UNSPECIFIED,
		UserID: 1,
	})
	if err == nil {
		t.Fatalf("expected error for unsupported type")
	}
}

func TestGetDailyMetricsValidation(t *testing.T) {
	svc := New(&stubRepo{})

	_, err := svc.GetDailyMetrics(context.Background(), ports.DailyMetricsRequest{})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument error, got %v", err)
	}
}

func TestGetDailyMetricsDefaultsDate(t *testing.T) {
	repo := &stubRepo{
		metrics: &entities.DailyTaskMetrics{UserID: 7},
	}
	svc := New(repo)

	metrics, err := svc.GetDailyMetrics(context.Background(), ports.DailyMetricsRequest{UserID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.UserID != 7 {
		t.Fatalf("unexpected metrics user id: %d", metrics.UserID)
	}
	if repo.lastDate.IsZero() {
		t.Fatalf("expected date to be defaulted")
	}
}

func TestBuildDelta(t *testing.T) {
	tests := []struct {
		name    string
		typ     analyticsv1.TaskEventType
		expect  ports.MetricsDelta
		wantErr bool
	}{
		{name: "created", typ: analyticsv1.TaskEventType_TASK_EVENT_TYPE_CREATED, expect: ports.MetricsDelta{Created: 1, Total: 1}},
		{name: "completed", typ: analyticsv1.TaskEventType_TASK_EVENT_TYPE_COMPLETED, expect: ports.MetricsDelta{Completed: 1}},
		{name: "deleted", typ: analyticsv1.TaskEventType_TASK_EVENT_TYPE_DELETED, expect: ports.MetricsDelta{Total: -1}},
		{name: "unsupported", typ: analyticsv1.TaskEventType_TASK_EVENT_TYPE_UNSPECIFIED, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta, err := buildDelta(tt.typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if delta != tt.expect {
				t.Fatalf("expected %+v, got %+v", tt.expect, delta)
			}
		})
	}
}
