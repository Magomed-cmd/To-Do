package auth

import (
	"net/http"

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
	router.POST("/auth/register", h.Register)
	router.POST("/auth/login", h.Login)
	router.POST("/auth/refresh", h.Refresh)
	router.POST("/auth/logout", h.Logout)
	router.POST("/auth/validate", h.Validate)
}

func (h *Handler) Register(ctx *gin.Context) {
	var request dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	result, err := h.service.Register(ctx.Request.Context(), ports.RegisterInput{
		Email:    request.Email,
		Name:     request.Name,
		Password: request.Password,
	})
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, dto.NewAuthResponse(*result))
}

func (h *Handler) Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	result, err := h.service.Login(ctx.Request.Context(), ports.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewAuthResponse(*result))
}

func (h *Handler) Refresh(ctx *gin.Context) {
	var request dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	tokens, err := h.service.RefreshToken(ctx.Request.Context(), request.RefreshToken)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewTokensResponse(*tokens))
}

func (h *Handler) Logout(ctx *gin.Context) {
	var request dto.LogoutRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	if err := h.service.Logout(ctx.Request.Context(), request.RefreshToken); err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *Handler) Validate(ctx *gin.Context) {
	token := common.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "MISSING_TOKEN"})
		return
	}
	claims, err := h.tokens.ParseAccessToken(token)
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
