package ports

import "time"

type TokenPayload struct {
	UserID int64
	Email  string
	Role   string
}

type TokenClaims struct {
	UserID    int64
	Email     string
	Role      string
	TokenID   string
	ExpiresAt time.Time
}

type TokenManager interface {
	GenerateAccessToken(payload TokenPayload) (string, time.Time, error)
	GenerateRefreshToken(payload TokenPayload, tokenID string) (string, time.Time, error)
	ParseAccessToken(token string) (*TokenClaims, error)
	ParseRefreshToken(token string) (*TokenClaims, error)
}
