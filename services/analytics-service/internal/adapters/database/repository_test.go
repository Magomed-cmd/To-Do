package database

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v2"

	"todoapp/services/analytics-service/internal/domain/entities"
	"todoapp/services/analytics-service/internal/ports"
)

func TestUpdateTaskMetrics(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	now := time.Date(2024, 10, 1, 12, 30, 0, 0, time.UTC)
	normalized := normalizeDate(now)

	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO analytics_service.task_metrics (user_id, date, created_tasks, completed_tasks, total_tasks)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, date) DO UPDATE
SET created_tasks = analytics_service.task_metrics.created_tasks + EXCLUDED.created_tasks,
    completed_tasks = analytics_service.task_metrics.completed_tasks + EXCLUDED.completed_tasks,
    total_tasks = GREATEST(0, analytics_service.task_metrics.total_tasks + EXCLUDED.total_tasks),
    updated_at = NOW()
`)).
		WithArgs(int64(7), normalized, int32(1), int32(0), int32(2)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := repo.UpdateTaskMetrics(context.Background(), 7, now, ports.MetricsDelta{Created: 1, Total: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetDailyMetricsFound(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	date := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	normalized := normalizeDate(date)

	rows := pgxmock.NewRows([]string{"user_id", "date", "created_tasks", "completed_tasks", "total_tasks", "updated_at"}).
		AddRow(int64(5), normalized, int32(2), int32(1), int32(3), normalized.Add(time.Hour))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT user_id, date, created_tasks, completed_tasks, total_tasks, updated_at
FROM analytics_service.task_metrics
WHERE user_id = $1 AND date = $2
`)).
		WithArgs(int64(5), normalized).
		WillReturnRows(rows)

	metrics, err := repo.GetDailyMetrics(context.Background(), 5, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := &entities.DailyTaskMetrics{
		UserID:         5,
		Date:           normalized,
		CreatedTasks:   2,
		CompletedTasks: 1,
		TotalTasks:     3,
		UpdatedAt:      normalized.Add(time.Hour),
	}
	if *metrics != *want {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetDailyMetricsNoRows(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT user_id, date, created_tasks, completed_tasks, total_tasks, updated_at
FROM analytics_service.task_metrics
WHERE user_id = $1 AND date = $2
`)).
		WithArgs(int64(9), pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	metrics, err := repo.GetDailyMetrics(context.Background(), 9, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.UserID != 9 {
		t.Fatalf("unexpected user id: %d", metrics.UserID)
	}
	if metrics.Date.IsZero() {
		t.Fatalf("expected normalized date")
	}
}

func TestNormalizeDate(t *testing.T) {
	zero := normalizeDate(time.Time{})
	if zero.Hour() != 0 || zero.Minute() != 0 || zero.Second() != 0 || zero.Location() != time.UTC {
		t.Fatalf("expected zero date to be normalized to utc midnight, got %v", zero)
	}

	value := time.Date(2024, 2, 3, 15, 4, 5, 0, time.FixedZone("MSK", 3*60*60))
	normalized := normalizeDate(value)
	if normalized.Year() != 2024 || normalized.Month() != 2 || normalized.Day() != 3 {
		t.Fatalf("unexpected normalized date: %v", normalized)
	}
	if normalized.Location() != time.UTC {
		t.Fatalf("expected UTC location")
	}
}
