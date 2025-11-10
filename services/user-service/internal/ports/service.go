package ports

import (
	"context"
	"time"

	"todoapp/services/user-service/internal/domain/entities"
)

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthTokens struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type AuthResult struct {
	User   entities.User
	Tokens AuthTokens
}

type UserService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	GitHubLogin(ctx context.Context, input OAuthLoginInput) (*AuthResult, error)
	GetProfile(ctx context.Context, userID int64) (*entities.User, error)
	UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*entities.User, error)
	GetPreferences(ctx context.Context, userID int64) (*entities.UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID int64, input UpdatePreferencesInput) (*entities.UserPreferences, error)
	ListUsers(ctx context.Context, limit, offset int) ([]entities.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role string) (*entities.User, error)
	UpdateUserStatus(ctx context.Context, userID int64, isActive bool) (*entities.User, error)
}

type OAuthLoginInput struct {
	ProviderID int64
	Email      string
	Name       string
	AvatarURL  string
}

type UpdateProfileInput struct {
	Name      *string
	AvatarURL *string
}

type UpdatePreferencesInput struct {
	NotificationsEnabled *bool
	EmailNotifications   *bool
	Theme                *string
	Language             *string
	Timezone             *string
}
