package export

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/adapters/http/middleware"
	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
)

// Handler handles task export HTTP requests.
type Handler struct {
	service ports.TaskService
}

// New creates a new export handler.
func New(service ports.TaskService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers export routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/export/csv", h.ExportCSV)
	router.GET("/export/ical", h.ExportICal)
}

// ExportCSV exports tasks as CSV file.
func (h *Handler) ExportCSV(ctx *gin.Context) {
	h.export(ctx, entities.ExportFormatCSV)
}

// ExportICal exports tasks as iCalendar file.
func (h *Handler) ExportICal(ctx *gin.Context) {
	h.export(ctx, entities.ExportFormatICal)
}

func (h *Handler) export(ctx *gin.Context, format entities.ExportFormat) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	data, filename, err := h.service.ExportTasks(ctx.Request.Context(), claims.UserID, format)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	ctx.Data(http.StatusOK, format.ContentType(), data)
}
