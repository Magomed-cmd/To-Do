package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"todoapp/services/user-service/internal/domain/entities"

	"todoapp/services/user-service/internal/domain"
	"todoapp/services/user-service/internal/ports"
)

type contextKey string

const txContextKey contextKey = "pgx_tx"

var _ ports.UserRepository = (*PostgresUserRepository)(nil)

type Pool interface {
	querier
	Begin(ctx context.Context) (pgx.Tx, error)
}

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
}

type PostgresUserRepository struct {
	pool Pool
}

func NewPostgresUserRepository(pool Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

const baseSelect = `
SELECT
    u.id,
    u.email,
    u.github_id,
    u.name,
    u.avatar_url,
    u.role,
    u.is_active,
    u.password_hash,
    u.created_at,
    u.updated_at,
    COALESCE(p.notifications_enabled, false),
    COALESCE(p.email_notifications, false),
    COALESCE(p.theme, ''),
    COALESCE(p.language, ''),
    COALESCE(p.timezone, ''),
    COALESCE(p.updated_at, '1970-01-01')
FROM user_service.users u
LEFT JOIN user_service.user_preferences p ON p.user_id = u.id
`

func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	q := r.querier(ctx)
	query := `
INSERT INTO user_service.users (email, github_id, name, avatar_url, role, is_active, password_hash)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id, created_at, updated_at
`
	if err := q.QueryRow(ctx, query,
		user.Email,
		user.GitHubID,
		user.Name,
		user.AvatarURL,
		user.Role,
		user.IsActive,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return err
	}
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	q := r.querier(ctx)
	row := q.QueryRow(ctx, baseSelect+" WHERE u.id = $1", id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	q := r.querier(ctx)
	row := q.QueryRow(ctx, baseSelect+" WHERE u.email = $1", email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	q := r.querier(ctx)
	query := `
UPDATE user_service.users
SET email = $1,
    github_id = $2,
    name = $3,
    avatar_url = $4,
    role = $5,
    is_active = $6,
    password_hash = $7,
    updated_at = $8
WHERE id = $9
`
	tag, err := q.Exec(ctx, query,
		user.Email,
		user.GitHubID,
		user.Name,
		user.AvatarURL,
		user.Role,
		user.IsActive,
		user.PasswordHash,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
	q := r.querier(ctx)
	tag, err := q.Exec(ctx, `DELETE FROM user_service.users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpsertPreferences(ctx context.Context, prefs entities.UserPreferences) error {
	q := r.querier(ctx)
	query := `
INSERT INTO user_service.user_preferences (user_id, notifications_enabled, email_notifications, theme, language, timezone, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id) DO UPDATE SET
    notifications_enabled = EXCLUDED.notifications_enabled,
    email_notifications = EXCLUDED.email_notifications,
    theme = EXCLUDED.theme,
    language = EXCLUDED.language,
    timezone = EXCLUDED.timezone,
    updated_at = EXCLUDED.updated_at
`
	_, err := q.Exec(ctx, query,
		prefs.UserID,
		prefs.NotificationsEnabled,
		prefs.EmailNotifications,
		prefs.Theme,
		prefs.Language,
		prefs.Timezone,
		time.Now(),
	)
	return err
}

func (r *PostgresUserRepository) GetPreferences(ctx context.Context, userID int64) (*entities.UserPreferences, error) {
	q := r.querier(ctx)
	row := q.QueryRow(ctx, `
SELECT user_id, notifications_enabled, email_notifications, theme, language, timezone, updated_at
FROM user_service.user_preferences
WHERE user_id = $1
`, userID)
	var prefs entities.UserPreferences
	if err := row.Scan(&prefs.UserID, &prefs.NotificationsEnabled, &prefs.EmailNotifications, &prefs.Theme, &prefs.Language, &prefs.Timezone, &prefs.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &entities.UserPreferences{UserID: userID}, nil
		}
		return nil, err
	}
	return &prefs, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, limit, offset int) ([]entities.User, error) {
	q := r.querier(ctx)
	rows, err := q.Query(ctx, baseSelect+" ORDER BY u.id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresUserRepository) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return WithTransaction(ctx, r.pool, fn)
}

func (r *PostgresUserRepository) querier(ctx context.Context) querier {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func scanUser(row rowScanner) (*entities.User, error) {
	var (
		user         entities.User
		githubIDNull sql.NullInt64
		avatarNull   sql.NullString
		roleNull     sql.NullString
		prefTheme    sql.NullString
		prefLang     sql.NullString
		prefTZ       sql.NullString
		prefUpdated  time.Time
	)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&githubIDNull,
		&user.Name,
		&avatarNull,
		&roleNull,
		&user.IsActive,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Preferences.NotificationsEnabled,
		&user.Preferences.EmailNotifications,
		&prefTheme,
		&prefLang,
		&prefTZ,
		&prefUpdated,
	)
	if err != nil {
		return nil, err
	}
	if githubIDNull.Valid {
		user.GitHubID = &githubIDNull.Int64
	}
	if avatarNull.Valid {
		user.AvatarURL = avatarNull.String
	}
	if roleNull.Valid {
		user.Role = roleNull.String
	}
	if prefTheme.Valid {
		user.Preferences.Theme = prefTheme.String
	}
	if prefLang.Valid {
		user.Preferences.Language = prefLang.String
	}
	if prefTZ.Valid {
		user.Preferences.Timezone = prefTZ.String
	}
	user.Preferences.UserID = user.ID
	user.Preferences.UpdatedAt = prefUpdated
	return &user, nil
}

type rowScanner interface {
	Scan(dest ...any) error
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
