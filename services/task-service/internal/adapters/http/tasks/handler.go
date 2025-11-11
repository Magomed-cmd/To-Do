package tasks

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/adapters/http/common"
	"todoapp/services/task-service/internal/adapters/http/middleware"
	"todoapp/services/task-service/internal/dto"
	"todoapp/services/task-service/internal/ports"
)

type Handler struct {
	service ports.TaskService
}

func New(service ports.TaskService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/tasks", h.ListTasks)
	router.POST("/tasks", h.CreateTask)
	router.GET("/tasks/:id", h.GetTask)
	router.PUT("/tasks/:id", h.UpdateTask)
	router.PATCH("/tasks/:id/status", h.UpdateTaskStatus)
	router.DELETE("/tasks/:id", h.DeleteTask)

	router.GET("/tasks/:id/comments", h.ListComments)
	router.POST("/tasks/:id/comments", h.CreateComment)

	router.GET("/categories", h.ListCategories)
	router.POST("/categories", h.CreateCategory)
	router.DELETE("/categories/:id", h.DeleteCategory)
}

func (h *Handler) ListTasks(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	var filter dto.TaskFilterRequest
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	tasks, err := h.service.ListTasks(ctx.Request.Context(), claims.UserID, filter.ToFilter())
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewTaskResponses(tasks))
}

func (h *Handler) CreateTask(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	var request dto.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	task, err := h.service.CreateTask(ctx.Request.Context(), request.ToInput(claims.UserID))
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewTaskResponse(*task))
}

func (h *Handler) GetTask(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	taskID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	task, err := h.service.GetTask(ctx.Request.Context(), claims.UserID, taskID)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewTaskResponse(*task))
}

func (h *Handler) UpdateTask(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	taskID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	var request dto.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	task, err := h.service.UpdateTask(ctx.Request.Context(), request.ToInput(claims.UserID, taskID))
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewTaskResponse(*task))
}

func (h *Handler) UpdateTaskStatus(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	taskID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	var request dto.UpdateTaskStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	task, err := h.service.UpdateTaskStatus(ctx.Request.Context(), claims.UserID, taskID, request.ToStatus())
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewTaskResponse(*task))
}

func (h *Handler) DeleteTask(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	taskID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	if err := h.service.DeleteTask(ctx.Request.Context(), claims.UserID, taskID); err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *Handler) ListCategories(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	categories, err := h.service.ListCategories(ctx.Request.Context(), claims.UserID)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewCategoryResponses(categories))
}

func (h *Handler) CreateCategory(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	var request dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	category, err := h.service.CreateCategory(ctx.Request.Context(), request.ToInput(claims.UserID))
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewCategoryResponse(*category))
}

func (h *Handler) DeleteCategory(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	categoryID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	if err := h.service.DeleteCategory(ctx.Request.Context(), claims.UserID, categoryID); err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *Handler) ListComments(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	taskID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	comments, err := h.service.ListComments(ctx.Request.Context(), claims.UserID, taskID)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewCommentResponses(comments))
}

func (h *Handler) CreateComment(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	taskID, err := parseID(ctx.Param("id"))
	if err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	var request dto.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}

	comment, err := h.service.AddComment(ctx.Request.Context(), request.ToInput(claims.UserID, taskID))
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewCommentResponse(*comment))
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
