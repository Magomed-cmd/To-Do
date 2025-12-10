package app

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	metricshttp "todoapp/services/analytics-service/internal/adapters/http/metrics"
	"todoapp/services/analytics-service/internal/ports"
)

type HTTPDeps struct {
	ServiceName string
	Analytics   ports.AnalyticsService
}

func NewRouter(deps HTTPDeps) (*gin.Engine, error) {
	if deps.Analytics == nil {
		return nil, errors.New("analytics service is required")
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", healthHandler(deps.ServiceName))
	router.HEAD("/health", healthHandler(deps.ServiceName))

	handler := metricshttp.New(deps.Analytics)
	handler.RegisterRoutes(router)

	return router, nil
}

func healthHandler(serviceName string) gin.HandlerFunc {
	if serviceName == "" {
		serviceName = "analytics-service"
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
