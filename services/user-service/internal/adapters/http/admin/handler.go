package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todoapp/services/user-service/internal/adapters/http/common"
	"todoapp/services/user-service/internal/dto"
	"todoapp/services/user-service/internal/ports"
)

type Handler struct {
	service ports.UserService
}

func New(service ports.UserService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/admin/users", h.ListUsers)
	router.PUT("/admin/users/:id/role", h.UpdateRole)
	router.PUT("/admin/users/:id/status", h.UpdateStatus)
}

func (h *Handler) ListUsers(ctx *gin.Context) {
	limit, offset := parsePagination(ctx)
	users, err := h.service.ListUsers(ctx.Request.Context(), limit, offset)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, dto.NewUserResponse(user))
	}
	ctx.JSON(http.StatusOK, responses)
}

func (h *Handler) UpdateRole(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}
	var request dto.UpdateUserRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	user, err := h.service.UpdateUserRole(ctx.Request.Context(), id, request.Role)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserResponse(*user))
}

func (h *Handler) UpdateStatus(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}
	var request dto.UpdateUserStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	user, err := h.service.UpdateUserStatus(ctx.Request.Context(), id, request.IsActive)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserResponse(*user))
}

func parsePagination(ctx *gin.Context) (int, int) {
	limit := 20
	offset := 0
	if value := ctx.Query("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if value := ctx.Query("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
