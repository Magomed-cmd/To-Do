package entities

import "time"

type User struct {
	ID           int64
	Email        string
	GitHubID     *int64
	Name         string
	AvatarURL    string
	Role         string
	IsActive     bool
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Preferences  UserPreferences
	Sessions     []UserSession
}

type UserSession struct {
	ID           int64
	UserID       int64
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type UserPreferences struct {
	UserID               int64
	NotificationsEnabled bool
	EmailNotifications   bool
	Theme                string
	Language             string
	Timezone             string
	UpdatedAt            time.Time
}
