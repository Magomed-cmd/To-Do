package database

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
	"todoapp/services/user-service/internal/domain/entities"

	"todoapp/services/user-service/internal/domain"
)

func TestCreate(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	user := &entities.User{
		Email:        "test@example.com",
		Name:         "Test",
		Role:         "user",
		IsActive:     true,
		PasswordHash: "hash",
	}

	mock.ExpectQuery(`INSERT INTO user_service.users`).
		WithArgs(user.Email, user.GitHubID, user.Name, user.AvatarURL, user.Role, user.IsActive, user.PasswordHash).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), time.Now(), time.Now()))

	require.NoError(t, repo.Create(context.Background(), user))
	require.Equal(t, int64(1), user.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByIDNotFound(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	mock.ExpectQuery(regexp.QuoteMeta(baseSelect + " WHERE u.id = $1")).WithArgs(int64(1)).WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), 1)
	require.Equal(t, domain.ErrUserNotFound, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(baseSelect + " WHERE u.email = $1")).
		WithArgs("test@example.com").
		WillReturnRows(pgxmock.NewRows([]string{
			"id",
			"email",
			"github_id",
			"name",
			"avatar_url",
			"role",
			"is_active",
			"password_hash",
			"created_at",
			"updated_at",
			"notifications_enabled",
			"email_notifications",
			"theme",
			"language",
			"timezone",
			"updated_at",
		}).AddRow(
			int64(1),
			"test@example.com",
			sql.NullInt64{Int64: 99, Valid: true},
			"Test",
			sql.NullString{String: "https://", Valid: true},
			sql.NullString{String: "user", Valid: true},
			true,
			"hash",
			now,
			now,
			true,
			true,
			sql.NullString{String: "dark", Valid: true},
			sql.NullString{String: "ru", Valid: true},
			sql.NullString{String: "Europe/Moscow", Valid: true},
			now,
		))

	user, err := repo.GetByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, user.GitHubID)
	require.Equal(t, int64(99), *user.GitHubID)
	require.Equal(t, "dark", user.Preferences.Theme)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateNotFound(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	user := &entities.User{ID: 42}
	mock.ExpectExec(`UPDATE user_service.users`).
		WithArgs(user.Email, user.GitHubID, user.Name, user.AvatarURL, user.Role, user.IsActive, user.PasswordHash, pgxmock.AnyArg(), user.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = repo.Update(context.Background(), user)
	require.Equal(t, domain.ErrUserNotFound, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	mock.ExpectExec(`DELETE FROM user_service.users`).WithArgs(int64(10)).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	require.NoError(t, repo.Delete(context.Background(), 10))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertPreferences(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	prefs := entities.UserPreferences{UserID: 1, NotificationsEnabled: true, EmailNotifications: true, Theme: "dark", Language: "en", Timezone: "UTC"}
	mock.ExpectExec(`INSERT INTO user_service.user_preferences`).
		WithArgs(prefs.UserID, prefs.NotificationsEnabled, prefs.EmailNotifications, prefs.Theme, prefs.Language, prefs.Timezone, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	require.NoError(t, repo.UpsertPreferences(context.Background(), prefs))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTransaction(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()
	mock := pool

	repo := NewPostgresUserRepository(pool)
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = repo.WithTransaction(context.Background(), func(ctx context.Context) error {
		require.NotNil(t, TxFromContext(ctx))
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
