package profile

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

	"todoapp/services/user-service/internal/adapters/http/middleware"
	"todoapp/services/user-service/internal/ports"
)

func TestGetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	user := &entities.User{ID: 1, Email: "user@example.com"}
	mockSvc.On("GetProfile", mockCtx(), int64(1)).Return(user, nil)
	handler := New(mockSvc)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(middleware.ContextUserClaimsKey, &ports.TokenClaims{UserID: 1})
	})
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mockSvc.AssertExpectations(t)
}

func TestUpdatePreferences(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	prefs := &entities.UserPreferences{UserID: 1, Theme: "dark", UpdatedAt: time.Now()}
	mockSvc.On("UpdatePreferences", mockCtx(), int64(1), ports.UpdatePreferencesInput{Theme: strPtr("dark")}).Return(prefs, nil)
	handler := New(mockSvc)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(middleware.ContextUserClaimsKey, &ports.TokenClaims{UserID: 1})
	})
	handler.RegisterRoutes(router)
	body, _ := json.Marshal(map[string]any{"theme": "dark"})
	req := httptest.NewRequest(http.MethodPut, "/users/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mockSvc.AssertExpectations(t)
}

func strPtr(v string) *string {
	return &v
}

func mockCtx() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil })
}
