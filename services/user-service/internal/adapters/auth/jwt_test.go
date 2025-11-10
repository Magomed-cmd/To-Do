package auth

import (
	"testing"
	"time"

	"todoapp/services/user-service/internal/ports"
)

func TestGenerateAndParseTokens(t *testing.T) {
	manager := NewJWTManager("access", "refresh", time.Minute, time.Hour)
	fixed := time.Now().Add(time.Hour)
	manager.WithNow(func() time.Time { return fixed })
	payload := ports.TokenPayload{UserID: 10, Email: "user@example.com", Role: "user"}

	access, accessExp, err := manager.GenerateAccessToken(payload)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if accessExp != fixed.Add(time.Minute) {
		t.Fatalf("unexpected expiration")
	}
	claims, err := manager.ParseAccessToken(access)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if claims.UserID != 10 || claims.Email != "user@example.com" {
		t.Fatalf("unexpected claims")
	}

	refresh, refreshExp, err := manager.GenerateRefreshToken(payload, "tid")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if refreshExp != fixed.Add(time.Hour) {
		t.Fatalf("unexpected refresh expiration")
	}
	refreshClaims, err := manager.ParseRefreshToken(refresh)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if refreshClaims.TokenID != "tid" {
		t.Fatalf("expected token id")
	}
}

func TestParseInvalidToken(t *testing.T) {
	manager := NewJWTManager("access", "refresh", time.Minute, time.Hour)
	_, err := manager.ParseAccessToken("invalid.token")
	if err == nil {
		t.Fatalf("expected error")
	}
}
