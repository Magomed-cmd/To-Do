package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"todoapp/services/user-service/internal/domain/entities"

	"todoapp/services/user-service/internal/domain"
	"todoapp/services/user-service/internal/ports"
)

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	mockTokens := NewMockTokenManager(t)
	result := &ports.AuthResult{User: entities.User{ID: 1, Email: "user@example.com"}}
	mockSvc.On("Register", mockCtx(), ports.RegisterInput{Email: "user@example.com", Name: "User", Password: "password123"}).Return(result, nil)
	handler := New(mockSvc, mockTokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	body := []byte(`{"email":"user@example.com","name":"User","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusCreated, res.Code)
	mockSvc.AssertExpectations(t)
}

func TestLoginInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	mockTokens := NewMockTokenManager(t)
	mockSvc.On("Login", mockCtx(), ports.LoginInput{Email: "user@example.com", Password: "wrongpass"}).Return(nil, domain.ErrInvalidCredentials)
	handler := New(mockSvc, mockTokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	body := []byte(`{"email":"user@example.com","password":"wrongpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	mockSvc.AssertExpectations(t)
}

func TestRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	mockTokens := NewMockTokenManager(t)
	tokens := &ports.AuthTokens{AccessToken: "a", RefreshToken: "r"}
	mockSvc.On("RefreshToken", mockCtx(), "refresh").Return(tokens, nil)
	handler := New(mockSvc, mockTokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	body := []byte(`{"refreshToken":"refresh"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mockSvc.AssertExpectations(t)
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	mockTokens := NewMockTokenManager(t)
	mockSvc.On("Logout", mockCtx(), "refresh").Return(nil)
	handler := New(mockSvc, mockTokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	body := []byte(`{"refreshToken":"refresh"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNoContent, res.Code)
	mockSvc.AssertExpectations(t)
}

func TestValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	mockTokens := NewMockTokenManager(t)
	claims := &ports.TokenClaims{UserID: 1, Email: "user@example.com", Role: "user", ExpiresAt: time.Now().Add(time.Hour)}
	mockTokens.On("ParseAccessToken", "token").Return(claims, nil)
	handler := New(mockSvc, mockTokens)
	router := gin.New()
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodPost, "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mockTokens.AssertExpectations(t)
}

func mockCtx() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil })
}
