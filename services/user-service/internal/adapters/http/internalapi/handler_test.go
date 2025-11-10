package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"todoapp/services/user-service/internal/domain/entities"

	"todoapp/services/user-service/internal/ports"
)

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := NewMockUserService(t)
	service.On("GetProfile", mockCtx(), int64(1)).Return(&entities.User{ID: 1}, nil)
	tokens := NewMockTokenManager(t)
	handler := New(service, tokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodGet, "/internal/users/1", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	service.AssertExpectations(t)
}

func TestValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := NewMockUserService(t)
	tokens := NewMockTokenManager(t)
	claims := &ports.TokenClaims{UserID: 1, Email: "user@example.com", Role: "user", ExpiresAt: time.Now()}
	tokens.On("ParseAccessToken", "token").Return(claims, nil)
	handler := New(service, tokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	body, _ := json.Marshal(map[string]any{"token": "token"})
	req := httptest.NewRequest(http.MethodPost, "/internal/auth/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	tokens.AssertExpectations(t)
}

func mockCtx() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil })
}
