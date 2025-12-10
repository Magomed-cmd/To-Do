package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/ports"
)

type tokenStub struct {
	claims *ports.TokenClaims
	err    error
}

func (t tokenStub) GenerateAccessToken(payload ports.TokenPayload) (string, time.Time, error) {
	return "", time.Time{}, nil
}
func (t tokenStub) GenerateRefreshToken(payload ports.TokenPayload, tokenID string) (string, time.Time, error) {
	return "", time.Time{}, nil
}
func (t tokenStub) ParseAccessToken(token string) (*ports.TokenClaims, error) { return t.claims, t.err }
func (t tokenStub) ParseRefreshToken(token string) (*ports.TokenClaims, error) {
	return t.claims, t.err
}

func TestJWTMiddlewareValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		header     string
		tokenErr   error
		wantStatus int
	}{
		{name: "missing header", header: "", wantStatus: http.StatusUnauthorized},
		{name: "invalid header", header: "Token abc", wantStatus: http.StatusUnauthorized},
		{name: "bad token", header: "Bearer abc", tokenErr: errors.New("fail"), wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		ctx, r := gin.CreateTestContext(rec)
		r.Use(New(tokenStub{err: tt.tokenErr}).JWT())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if tt.header != "" {
			req.Header.Set("Authorization", tt.header)
		}
		r.ServeHTTP(rec, req)
		if rec.Code != tt.wantStatus {
			t.Fatalf("%s: expected %d, got %d", tt.name, tt.wantStatus, rec.Code)
		}
		_ = ctx
	}
}

func TestJWTMiddlewareSuccessAndCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	claims := &ports.TokenClaims{UserID: 1}

	router := gin.New()
	router.Use(New(tokenStub{claims: claims}).JWT())
	router.GET("/", func(c *gin.Context) {
		if got, ok := CurrentUser(c); !ok || got.UserID != claims.UserID {
			t.Fatalf("claims not stored in context")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
