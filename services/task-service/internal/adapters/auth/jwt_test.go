package auth

import (
	"testing"
	"time"

	"todoapp/services/task-service/internal/ports"
)

func TestAccessTokenGenerationAndParsing(t *testing.T) {
	mgr := NewJWTManager("access", "refresh", time.Minute, time.Hour)
	now := time.Now().UTC()
	mgr.WithNow(func() time.Time { return now })

	payload := ports.TokenPayload{UserID: 7, Email: "user@example.com", Role: "user"}
	token, exp, err := mgr.GenerateAccessToken(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exp.Equal(now.Add(time.Minute)) {
		t.Fatalf("unexpected expiration: %v", exp)
	}

	claims, err := mgr.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if claims.UserID != payload.UserID || claims.Email != payload.Email || claims.Role != payload.Role {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestRefreshTokenGenerationAndParsing(t *testing.T) {
	mgr := NewJWTManager("access", "refresh", time.Minute, time.Hour)
	now := time.Now().UTC()
	mgr.WithNow(func() time.Time { return now })

	token, exp, err := mgr.GenerateRefreshToken(ports.TokenPayload{UserID: 3, Email: "e", Role: "admin"}, "token-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.IsZero() {
		t.Fatalf("expected expiration to be set")
	}

	claims, err := mgr.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to parse refresh token: %v", err)
	}
	if claims.TokenID != "token-id" || claims.UserID != 3 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseFailsOnWrongSecret(t *testing.T) {
	mgr := NewJWTManager("access", "refresh", time.Minute, time.Hour)
	now := time.Now().UTC()
	mgr.WithNow(func() time.Time { return now })

	token, _, _ := mgr.GenerateAccessToken(ports.TokenPayload{UserID: 1})
	mgr.accessSecret = []byte("other-secret")

	if _, err := mgr.ParseAccessToken(token); err == nil {
		t.Fatalf("expected parse error with wrong secret")
	}
}
