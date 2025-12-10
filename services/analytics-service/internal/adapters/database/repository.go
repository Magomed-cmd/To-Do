package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"todoapp/services/analytics-service/internal/domain/entities"
	"todoapp/services/analytics-service/internal/ports"
)

type PostgresRepository struct {
	pool pgxPool
}

type pgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewPostgresRepository(pool pgxPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) UpdateTaskMetrics(ctx context.Context, userID int64, date time.Time, delta ports.MetricsDelta) error {
	normalized := normalizeDate(date)

	const query = `
INSERT INTO analytics_service.task_metrics (user_id, date, created_tasks, completed_tasks, total_tasks)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, date) DO UPDATE
SET created_tasks = analytics_service.task_metrics.created_tasks + EXCLUDED.created_tasks,
    completed_tasks = analytics_service.task_metrics.completed_tasks + EXCLUDED.completed_tasks,
    total_tasks = GREATEST(0, analytics_service.task_metrics.total_tasks + EXCLUDED.total_tasks),
    updated_at = NOW()
`

	_, err := r.pool.Exec(ctx, query,
		userID,
		normalized,
		delta.Created,
		delta.Completed,
		delta.Total,
	)
	return err
}

func (r *PostgresRepository) GetDailyMetrics(ctx context.Context, userID int64, date time.Time) (*entities.DailyTaskMetrics, error) {
	normalized := normalizeDate(date)

	const query = `
SELECT user_id, date, created_tasks, completed_tasks, total_tasks, updated_at
FROM analytics_service.task_metrics
WHERE user_id = $1 AND date = $2
`

	row := r.pool.QueryRow(ctx, query, userID, normalized)

	var metrics entities.DailyTaskMetrics
	if err := row.Scan(&metrics.UserID, &metrics.Date, &metrics.CreatedTasks, &metrics.CompletedTasks, &metrics.TotalTasks, &metrics.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return &entities.DailyTaskMetrics{UserID: userID, Date: normalized}, nil
		}
		return nil, err
	}

	return &metrics, nil
}

func normalizeDate(value time.Time) time.Time {
	if value.IsZero() {
		value = time.Now()
	}
	y, m, d := value.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
