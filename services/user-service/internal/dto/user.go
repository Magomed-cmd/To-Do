package dto

import (
	"time"

	"todoapp/services/user-service/internal/domain/entities"
	"todoapp/services/user-service/internal/ports"
)

type CreateUserRequest struct {
	Email     string  `json:"email" mapstructure:"email" binding:"required,email,max=255"`
	Name      string  `json:"name" mapstructure:"name" binding:"required,min=2,max=255"`
	AvatarURL *string `json:"avatarUrl,omitempty" mapstructure:"avatarUrl" binding:"omitempty,url,max=500"`
	GitHubID  *int64  `json:"githubId,omitempty" mapstructure:"githubId" binding:"omitempty,gte=1"`
	Role      *string `json:"role,omitempty" mapstructure:"role" binding:"omitempty,oneof=user admin,max=50"`
	Password  string  `json:"password" binding:"required,min=8,max=128"`
}

type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty" mapstructure:"name" binding:"omitempty,min=2,max=255"`
	AvatarURL *string `json:"avatarUrl,omitempty" mapstructure:"avatarUrl" binding:"omitempty,url,max=500"`
	Role      *string `json:"role,omitempty" mapstructure:"role" binding:"omitempty,oneof=user admin,max=50"`
	IsActive  *bool   `json:"isActive,omitempty" mapstructure:"isActive"`
}

type UserResponse struct {
	ID          int64                   `json:"id" mapstructure:"id"`
	Email       string                  `json:"email" mapstructure:"email"`
	Name        string                  `json:"name" mapstructure:"name"`
	AvatarURL   string                  `json:"avatarUrl,omitempty" mapstructure:"avatarUrl"`
	Role        string                  `json:"role" mapstructure:"role"`
	IsActive    bool                    `json:"isActive" mapstructure:"isActive"`
	GitHubID    *int64                  `json:"githubId,omitempty" mapstructure:"githubId"`
	CreatedAt   time.Time               `json:"createdAt" mapstructure:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt" mapstructure:"updatedAt"`
	Preferences UserPreferencesResponse `json:"preferences" mapstructure:"preferences"`
	Sessions    []UserSessionResponse   `json:"sessions,omitempty" mapstructure:"sessions"`
}

type UserPreferencesResponse struct {
	NotificationsEnabled bool      `json:"notificationsEnabled" mapstructure:"notificationsEnabled"`
	EmailNotifications   bool      `json:"emailNotifications" mapstructure:"emailNotifications"`
	Theme                string    `json:"theme" mapstructure:"theme"`
	Language             string    `json:"language" mapstructure:"language"`
	Timezone             string    `json:"timezone" mapstructure:"timezone"`
	UpdatedAt            time.Time `json:"updatedAt" mapstructure:"updatedAt"`
}

type UserSessionResponse struct {
	ID           int64     `json:"id" mapstructure:"id"`
	RefreshToken string    `json:"refreshToken" mapstructure:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt" mapstructure:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt" mapstructure:"createdAt"`
}

type UpdateProfileRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=2,max=255"`
	AvatarURL *string `json:"avatarUrl" binding:"omitempty,url,max=500"`
}

type UpdatePreferencesRequest struct {
	NotificationsEnabled *bool   `json:"notificationsEnabled" binding:"omitempty"`
	EmailNotifications   *bool   `json:"emailNotifications" binding:"omitempty"`
	Theme                *string `json:"theme" binding:"omitempty,max=20"`
	Language             *string `json:"language" binding:"omitempty,len=2"`
	Timezone             *string `json:"timezone" binding:"omitempty,max=50"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"isActive" binding:"required"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type TokensResponse struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

type AuthResponse struct {
	User   UserResponse   `json:"user"`
	Tokens TokensResponse `json:"tokens"`
}

func NewUserResponse(u entities.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		IsActive:  u.IsActive,
		GitHubID:  u.GitHubID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Preferences: UserPreferencesResponse{
			NotificationsEnabled: u.Preferences.NotificationsEnabled,
			EmailNotifications:   u.Preferences.EmailNotifications,
			Theme:                u.Preferences.Theme,
			Language:             u.Preferences.Language,
			Timezone:             u.Preferences.Timezone,
			UpdatedAt:            u.Preferences.UpdatedAt,
		},
		Sessions: mapSessions(u.Sessions),
	}
}

func mapSessions(sessions []entities.UserSession) []UserSessionResponse {
	if len(sessions) == 0 {
		return nil
	}

	result := make([]UserSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, UserSessionResponse{
			ID:           session.ID,
			RefreshToken: session.RefreshToken,
			ExpiresAt:    session.ExpiresAt,
			CreatedAt:    session.CreatedAt,
		})
	}

	return result
}

func NewAuthResponse(result ports.AuthResult) AuthResponse {
	return AuthResponse{
		User:   NewUserResponse(result.User),
		Tokens: NewTokensResponse(result.Tokens),
	}
}

func NewTokensResponse(tokens ports.AuthTokens) TokensResponse {
	return TokensResponse{
		AccessToken:           tokens.AccessToken,
		AccessTokenExpiresAt:  tokens.AccessTokenExpiresAt,
		RefreshToken:          tokens.RefreshToken,
		RefreshTokenExpiresAt: tokens.RefreshTokenExpiresAt,
	}
}

func NewUserPreferencesResponse(prefs entities.UserPreferences) UserPreferencesResponse {
	return UserPreferencesResponse{
		NotificationsEnabled: prefs.NotificationsEnabled,
		EmailNotifications:   prefs.EmailNotifications,
		Theme:                prefs.Theme,
		Language:             prefs.Language,
		Timezone:             prefs.Timezone,
		UpdatedAt:            prefs.UpdatedAt,
	}
}
