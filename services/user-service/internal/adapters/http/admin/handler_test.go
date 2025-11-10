package admin

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
)

func TestListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	users := []entities.User{{ID: 1, Email: "user@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	mockSvc.On("ListUsers", mockCtx(), 5, 0).Return(users, nil)
	handler := New(mockSvc)
	router := gin.New()
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodGet, "/admin/users?limit=5", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	var payload []map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	assert.Len(t, payload, 1)
	mockSvc.AssertExpectations(t)
}

func TestUpdateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	user := &entities.User{ID: 1, Role: "admin"}
	mockSvc.On("UpdateUserRole", mockCtx(), int64(1), "admin").Return(user, nil)
	handler := New(mockSvc)
	router := gin.New()
	handler.RegisterRoutes(router)
	body := []byte(`{"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mockSvc.AssertExpectations(t)
}

func TestUpdateStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := NewMockUserService(t)
	user := &entities.User{ID: 1, IsActive: true}
	mockSvc.On("UpdateUserStatus", mockCtx(), int64(1), true).Return(user, nil)
	handler := New(mockSvc)
	router := gin.New()
	handler.RegisterRoutes(router)
	body := []byte(`{"isActive":true}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mockSvc.AssertExpectations(t)
}

func mockCtx() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil })
}
