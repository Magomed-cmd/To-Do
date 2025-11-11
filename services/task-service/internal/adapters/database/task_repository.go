package database

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"todoapp/services/task-service/internal/domain"
	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
)

type contextKey string

const txContextKey contextKey = "taskService.tx"

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Pool interface {
	querier
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresTaskRepository struct {
	pool Pool
}

func NewPostgresTaskRepository(pool Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{pool: pool}
}

var _ ports.TaskRepository = (*PostgresTaskRepository)(nil)

func (r *PostgresTaskRepository) CreateTask(ctx context.Context, task *entities.Task) error {
	const query = `
INSERT INTO task_service.tasks (
    user_id,
    title,
    description,
    status,
    priority,
    due_date,
    category_id
) VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id, created_at, updated_at
`

	q := r.querier(ctx)

	if err := q.QueryRow(ctx, query,
		task.UserID,
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Priority),
		task.DueDate,
		task.CategoryID,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (r *PostgresTaskRepository) UpdateTask(ctx context.Context, task *entities.Task) error {
	const query = `
UPDATE task_service.tasks
SET title = $1,
    description = $2,
    status = $3,
    priority = $4,
    due_date = $5,
    category_id = $6,
    updated_at = NOW()
WHERE id = $7
  AND user_id = $8
RETURNING updated_at
`

	q := r.querier(ctx)

	if err := q.QueryRow(ctx, query,
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Priority),
		task.DueDate,
		task.CategoryID,
		task.ID,
		task.UserID,
	).Scan(&task.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTaskNotFound
		}
		return err
	}

	return nil
}

func (r *PostgresTaskRepository) SoftDeleteTask(ctx context.Context, userID, taskID int64, deletedAt time.Time) error {
	const query = `
DELETE FROM task_service.tasks
WHERE id = $1
  AND user_id = $2
`

	q := r.querier(ctx)

	tag, err := q.Exec(ctx, query, taskID, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (r *PostgresTaskRepository) GetTask(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	q := r.querier(ctx)

	row := q.QueryRow(ctx, baseTaskSelect()+`
WHERE t.id = $1
  AND t.user_id = $2
`, taskID, userID)

	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}

	return task, nil
}

func (r *PostgresTaskRepository) ListTasks(ctx context.Context, userID int64, filter ports.TaskFilter) ([]entities.Task, error) {
	var (
		args      []any
		clauses   []string
		argsIndex = 1
	)

	clauses = append(clauses, "t.user_id = $"+itoa(argsIndex))
	args = append(args, userID)
	argsIndex++

	if len(filter.Statuses) > 0 {
		clauses = append(clauses, "t.status = ANY($"+itoa(argsIndex)+")")
		args = append(args, statusesToTextArray(filter.Statuses))
		argsIndex++
	}

	if len(filter.Priorities) > 0 {
		clauses = append(clauses, "t.priority = ANY($"+itoa(argsIndex)+")")
		args = append(args, prioritiesToTextArray(filter.Priorities))
		argsIndex++
	}

	if filter.CategoryID != nil {
		clauses = append(clauses, "t.category_id = $"+itoa(argsIndex))
		args = append(args, *filter.CategoryID)
		argsIndex++
	}

	if filter.Search != "" {
		search := "%" + strings.ToLower(filter.Search) + "%"
		clauses = append(clauses, "(LOWER(t.title) LIKE $"+itoa(argsIndex)+" OR LOWER(t.description) LIKE $"+itoa(argsIndex)+")")
		args = append(args, search)
		argsIndex++
	}

	if filter.DueFrom != nil {
		clauses = append(clauses, "t.due_date >= $"+itoa(argsIndex))
		args = append(args, *filter.DueFrom)
		argsIndex++
	}

	if filter.DueTo != nil {
		clauses = append(clauses, "t.due_date <= $"+itoa(argsIndex))
		args = append(args, *filter.DueTo)
		argsIndex++
	}

	query := baseTaskSelect() + "\nWHERE " + strings.Join(clauses, " AND ") + "\nORDER BY COALESCE(t.due_date, t.created_at) ASC\nLIMIT $" + itoa(argsIndex) + "\nOFFSET $" + itoa(argsIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	q := r.querier(ctx)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entities.Task

	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *PostgresTaskRepository) CreateCategory(ctx context.Context, category *entities.Category) error {
	const query = `
INSERT INTO task_service.categories (user_id, name)
VALUES ($1,$2)
RETURNING id, created_at
`

	q := r.querier(ctx)

	if err := q.QueryRow(ctx, query,
		category.UserID,
		category.Name,
	).Scan(&category.ID, &category.CreatedAt); err != nil {
		return err
	}

	category.UpdatedAt = category.CreatedAt

	return nil
}

func (r *PostgresTaskRepository) ListCategories(ctx context.Context, userID int64) ([]entities.Category, error) {
	const query = `
SELECT id, user_id, name, created_at
FROM task_service.categories
WHERE user_id = $1
ORDER BY name ASC
`

	q := r.querier(ctx)

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entities.Category

	for rows.Next() {
		var category entities.Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.CreatedAt); err != nil {
			return nil, err
		}
		category.UpdatedAt = category.CreatedAt
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *PostgresTaskRepository) GetCategory(ctx context.Context, userID, categoryID int64) (*entities.Category, error) {
	const query = `
SELECT id, user_id, name, created_at
FROM task_service.categories
WHERE id = $1
  AND user_id = $2
`

	q := r.querier(ctx)

	var category entities.Category

	if err := q.QueryRow(ctx, query, categoryID, userID).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}

	category.UpdatedAt = category.CreatedAt
	return &category, nil
}

func (r *PostgresTaskRepository) DeleteCategory(ctx context.Context, userID, categoryID int64) error {
	const query = `
DELETE FROM task_service.categories
WHERE id = $1
  AND user_id = $2
`

	q := r.querier(ctx)

	tag, err := q.Exec(ctx, query, categoryID, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

func (r *PostgresTaskRepository) CreateComment(ctx context.Context, comment *entities.TaskComment) error {
	const query = `
INSERT INTO task_service.task_comments (task_id, user_id, content)
VALUES ($1,$2,$3)
RETURNING id, created_at
`

	q := r.querier(ctx)

	if err := q.QueryRow(ctx, query,
		comment.TaskID,
		comment.UserID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt); err != nil {
		return err
	}

	return nil
}

func (r *PostgresTaskRepository) ListComments(ctx context.Context, userID, taskID int64) ([]entities.TaskComment, error) {
	const query = `
SELECT id, task_id, user_id, content, created_at
FROM task_service.task_comments
WHERE task_id = $1
  AND user_id = $2
ORDER BY created_at ASC
`

	q := r.querier(ctx)

	rows, err := q.Query(ctx, query, taskID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []entities.TaskComment

	for rows.Next() {
		var comment entities.TaskComment
		if err := rows.Scan(&comment.ID, &comment.TaskID, &comment.UserID, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *PostgresTaskRepository) querier(ctx context.Context) querier {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func TxFromContext(ctx context.Context) pgx.Tx {
	if ctx == nil {
		return nil
	}

	tx, _ := ctx.Value(txContextKey).(pgx.Tx)
	return tx
}

func WithTransaction(ctx context.Context, pool Pool, fn func(context.Context) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txContextKey, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(txCtx)
		return err
	}

	return tx.Commit(txCtx)
}

func baseTaskSelect() string {
	return `
SELECT
    t.id,
    t.user_id,
    t.title,
    t.description,
    t.status,
    t.priority,
    t.due_date,
    t.category_id,
    t.created_at,
    t.updated_at,
    c.id,
    c.user_id,
    c.name,
    c.created_at
FROM task_service.tasks t
LEFT JOIN task_service.categories c ON c.id = t.category_id
`
}

func scanTask(row rowScanner) (*entities.Task, error) {
	var (
		task            entities.Task
		dueDate         sql.NullTime
		categoryID      sql.NullInt64
		categoryEntity  sql.NullInt64
		categoryUserID  sql.NullInt64
		categoryName    sql.NullString
		categoryCreated sql.NullTime
	)

	if err := row.Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&dueDate,
		&categoryID,
		&task.CreatedAt,
		&task.UpdatedAt,
		&categoryEntity,
		&categoryUserID,
		&categoryName,
		&categoryCreated,
	); err != nil {
		return nil, err
	}

	if dueDate.Valid {
		value := dueDate.Time
		task.DueDate = &value
	}

	if categoryID.Valid {
		value := categoryID.Int64
		task.CategoryID = &value
	}

	if categoryEntity.Valid {
		category := entities.Category{
			ID:     categoryEntity.Int64,
			UserID: categoryUserID.Int64,
			Name:   categoryName.String,
		}
		if categoryCreated.Valid {
			category.CreatedAt = categoryCreated.Time
			category.UpdatedAt = categoryCreated.Time
		}
		task.Category = &category
	}

	return &task, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func statusesToTextArray(statuses []entities.TaskStatus) []string {
	out := make([]string, 0, len(statuses))
	for _, status := range statuses {
		out = append(out, string(status))
	}
	return out
}

func prioritiesToTextArray(priorities []entities.TaskPriority) []string {
	out := make([]string, 0, len(priorities))
	for _, priority := range priorities {
		out = append(out, string(priority))
	}
	return out
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
