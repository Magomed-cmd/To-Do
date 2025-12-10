package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"todoapp/services/analytics-service/internal/ports"
)

type Handler struct {
	service ports.AnalyticsService
}

func New(service ports.AnalyticsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/metrics/daily/:userId", h.GetDailyMetrics)
}

func (h *Handler) GetDailyMetrics(ctx *gin.Context) {
	userID, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil || userID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_USER_ID"})
		return
	}

	var date time.Time
	if raw := ctx.Query("date"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_DATE"})
			return
		}
		date = parsed
	}

	metrics, err := h.service.GetDailyMetrics(ctx.Request.Context(), ports.DailyMetricsRequest{UserID: userID, Date: date})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"userId":         metrics.UserID,
		"date":           metrics.Date.Format("2006-01-02"),
		"createdTasks":   metrics.CreatedTasks,
		"completedTasks": metrics.CompletedTasks,
		"totalTasks":     metrics.TotalTasks,
		"updatedAt":      metrics.UpdatedAt,
	})
}
