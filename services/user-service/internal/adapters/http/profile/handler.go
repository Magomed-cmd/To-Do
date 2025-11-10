package profile

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todoapp/services/user-service/internal/adapters/http/common"
	"todoapp/services/user-service/internal/adapters/http/middleware"
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
	router.GET("/users/profile", h.GetProfile)
	router.PUT("/users/profile", h.UpdateProfile)
	router.GET("/users/preferences", h.GetPreferences)
	router.PUT("/users/preferences", h.UpdatePreferences)
}

func (h *Handler) GetProfile(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}
	user, err := h.service.GetProfile(ctx.Request.Context(), claims.UserID)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserResponse(*user))
}

func (h *Handler) UpdateProfile(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}
	var request dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	user, err := h.service.UpdateProfile(ctx.Request.Context(), claims.UserID, ports.UpdateProfileInput{
		Name:      request.Name,
		AvatarURL: request.AvatarURL,
	})
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserResponse(*user))
}

func (h *Handler) GetPreferences(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}
	prefs, err := h.service.GetPreferences(ctx.Request.Context(), claims.UserID)
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserPreferencesResponse(*prefs))
}

func (h *Handler) UpdatePreferences(ctx *gin.Context) {
	claims, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}
	var request dto.UpdatePreferencesRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteValidationError(ctx, err)
		return
	}
	prefs, err := h.service.UpdatePreferences(ctx.Request.Context(), claims.UserID, ports.UpdatePreferencesInput{
		NotificationsEnabled: request.NotificationsEnabled,
		EmailNotifications:   request.EmailNotifications,
		Theme:                request.Theme,
		Language:             request.Language,
		Timezone:             request.Timezone,
	})
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.NewUserPreferencesResponse(*prefs))
}
