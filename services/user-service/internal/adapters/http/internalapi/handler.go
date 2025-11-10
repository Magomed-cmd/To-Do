package internalapi

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
	tokens  ports.TokenManager
}

func New(service ports.UserService, tokens ports.TokenManager) *Handler {
	return &Handler{service: service, tokens: tokens}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/internal/users/:id", h.GetUser)
	router.POST("/internal/auth/validate", h.ValidateToken)
}

func (h *Handler) GetUser(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}
	user, err := h.service.GetProfile(ctx.Request.Context(), id)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserResponse(*user))
}

func (h *Handler) ValidateToken(ctx *gin.Context) {
	var request dto.ValidateTokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	claims, err := h.tokens.ParseAccessToken(request.Token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_TOKEN"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"userId":    claims.UserID,
		"email":     claims.Email,
		"role":      claims.Role,
		"expiresAt": claims.ExpiresAt,
	})
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
