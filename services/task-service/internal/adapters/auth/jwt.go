package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"todoapp/services/task-service/internal/ports"
)

type tokenClaims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	now           func() time.Time
}

var _ ports.TokenManager = (*JWTManager)(nil)

func NewJWTManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		now:           time.Now,
	}
}

func (m *JWTManager) GenerateAccessToken(payload ports.TokenPayload) (string, time.Time, error) {
	expiresAt := m.now().Add(m.accessTTL)

	claims := tokenClaims{
		UserID: payload.UserID,
		Email:  payload.Email,
		Role:   payload.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(payload.UserID, 10),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(m.now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (m *JWTManager) GenerateRefreshToken(payload ports.TokenPayload, tokenID string) (string, time.Time, error) {
	expiresAt := m.now().Add(m.refreshTTL)

	claims := tokenClaims{
		UserID: payload.UserID,
		Email:  payload.Email,
		Role:   payload.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d:%s", payload.UserID, tokenID),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(m.now()),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.refreshSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (m *JWTManager) ParseAccessToken(token string) (*ports.TokenClaims, error) {
	claims, err := m.parse(token, m.accessSecret)
	if err != nil {
		return nil, err
	}

	return m.mapClaims(claims), nil
}

func (m *JWTManager) ParseRefreshToken(token string) (*ports.TokenClaims, error) {
	claims, err := m.parse(token, m.refreshSecret)
	if err != nil {
		return nil, err
	}

	return m.mapClaims(claims), nil
}

func (m *JWTManager) parse(token string, secret []byte) (*tokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*tokenClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) mapClaims(c *tokenClaims) *ports.TokenClaims {
	return &ports.TokenClaims{
		UserID:    c.UserID,
		Email:     c.Email,
		Role:      c.Role,
		TokenID:   c.RegisteredClaims.ID,
		ExpiresAt: c.RegisteredClaims.ExpiresAt.Time,
	}
}

func (m *JWTManager) WithNow(now func() time.Time) {
	m.now = now
}
