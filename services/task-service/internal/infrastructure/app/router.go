package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	exporthttp "todoapp/services/task-service/internal/adapters/http/export"
	middlewarehttp "todoapp/services/task-service/internal/adapters/http/middleware"
	taskshttp "todoapp/services/task-service/internal/adapters/http/tasks"
	"todoapp/services/task-service/internal/ports"
)

type HTTPDeps struct {
	TaskService ports.TaskService
	TokenMgr    ports.TokenManager
	ServiceName string
}

func NewRouter(deps HTTPDeps) (*gin.Engine, error) {
	if err := validateHTTPDeps(deps); err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", healthHandler(deps.ServiceName))
	router.HEAD("/health", healthHandler(deps.ServiceName))

	security := middlewarehttp.New(deps.TokenMgr)

	protected := router.Group("")
	protected.Use(security.JWT())

	taskHandler := taskshttp.New(deps.TaskService)
	taskHandler.RegisterRoutes(protected)

	exportHandler := exporthttp.New(deps.TaskService)
	exportHandler.RegisterRoutes(protected)

	return router, nil
}

func validateHTTPDeps(deps HTTPDeps) error {
	switch {
	case deps.TaskService == nil:
		return fmt.Errorf("task service is required")
	case deps.TokenMgr == nil:
		return fmt.Errorf("token manager is required")
	default:
		return nil
	}
}

func healthHandler(serviceName string) gin.HandlerFunc {
	if serviceName == "" {
		serviceName = "task-service"
	}

	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodHead {
			ctx.Status(http.StatusOK)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": serviceName,
		})
	}
}
