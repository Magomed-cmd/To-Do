package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"todoapp/services/user-service/internal/ports"
)

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mgr := NewMockTokenManager(t)
	mgr.On("ParseAccessToken", "token").Return(&ports.TokenClaims{UserID: 1, Role: "admin"}, nil)
	m := New(mgr)
	router := gin.New()
	router.Use(m.JWT())
	router.GET("/protected", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	mgr.AssertExpectations(t)
}

func TestRequireRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := New(NewMockTokenManager(t))
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(ContextUserClaimsKey, &ports.TokenClaims{Role: "admin"})
	})
	router.Use(m.RequireRoles("admin"))
	router.GET("/protected", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestRequireRolesForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := New(NewMockTokenManager(t))
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(ContextUserClaimsKey, &ports.TokenClaims{Role: "user"})
	})
	router.Use(m.RequireRoles("admin"))
	router.GET("/protected", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestExtractBearerToken(t *testing.T) {
	if extractBearerToken("Bearer token") != "token" {
		t.Fatalf("unexpected token")
	}
	if extractBearerToken("token") != "" {
		t.Fatalf("expected empty token")
	}
}
