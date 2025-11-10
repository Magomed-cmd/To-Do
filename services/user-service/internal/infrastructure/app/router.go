package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	authadapter "todoapp/services/user-service/internal/adapters/auth"
	adminhttp "todoapp/services/user-service/internal/adapters/http/admin"
	authhttp "todoapp/services/user-service/internal/adapters/http/auth"
	githubhttp "todoapp/services/user-service/internal/adapters/http/github"
	internalhttp "todoapp/services/user-service/internal/adapters/http/internalapi"
	middlewarehttp "todoapp/services/user-service/internal/adapters/http/middleware"
	profilehttp "todoapp/services/user-service/internal/adapters/http/profile"
	"todoapp/services/user-service/internal/ports"
)

type HTTPDeps struct {
	UserService ports.UserService
	TokenMgr    ports.TokenManager
	GitHubOAuth *authadapter.GitHubOAuth
}

func NewRouter(deps HTTPDeps) (*gin.Engine, error) {
	if err := validateHTTPDeps(deps); err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)

	security := middlewarehttp.New(deps.TokenMgr)

	protected := router.Group("")
	protected.Use(security.JWT())

	adminGroup := protected.Group("")
	adminGroup.Use(security.RequireRoles("admin"))

	authHandler := authhttp.New(deps.UserService, deps.TokenMgr)
	internalHandler := internalhttp.New(deps.UserService, deps.TokenMgr)
	githubHandler := githubhttp.New(deps.GitHubOAuth, deps.UserService)
	profileHandler := profilehttp.New(deps.UserService)
	adminHandler := adminhttp.New(deps.UserService)

	githubHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(router)
	internalHandler.RegisterRoutes(router)
	profileHandler.RegisterRoutes(protected)
	adminHandler.RegisterRoutes(adminGroup)

	return router, nil
}

func validateHTTPDeps(deps HTTPDeps) error {
	switch {
	case deps.UserService == nil:
		return fmt.Errorf("user service is required")
	case deps.TokenMgr == nil:
		return fmt.Errorf("token manager is required")
	case deps.GitHubOAuth == nil:
		return fmt.Errorf("github oauth provider is required")
	default:
		return nil
	}
}

func healthHandler(c *gin.Context) {
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "user-service",
	})
}
