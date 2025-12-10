package ports

import (
	"context"

	"todoapp/services/user-service/internal/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
	UpsertPreferences(ctx context.Context, prefs entities.UserPreferences) error
	GetPreferences(ctx context.Context, userID int64) (*entities.UserPreferences, error)
	List(ctx context.Context, limit, offset int) ([]entities.User, error)
	CreateSession(ctx context.Context, session entities.UserSession) error
	GetSession(ctx context.Context, token string) (*entities.UserSession, error)
	DeleteSession(ctx context.Context, token string) error
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}
