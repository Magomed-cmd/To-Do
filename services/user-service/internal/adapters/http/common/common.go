package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"todoapp/services/user-service/internal/domain"
)

func WriteValidationError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error":   "validation_error",
		"message": err.Error(),
	})
}

func WriteDomainError(ctx *gin.Context, err error) {
	if domainErr, ok := err.(*domain.DomainError); ok {
		switch domainErr {
		case domain.ErrUserAlreadyExists:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": domainErr.Code, "message": domainErr.Message})
		case domain.ErrUserNotFound:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": domainErr.Code, "message": domainErr.Message})
		case domain.ErrInvalidCredentials:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": domainErr.Code, "message": domainErr.Message})
		case domain.ErrUserInactive, domain.ErrUserLocked, domain.ErrUserSuspended:
			ctx.JSON(http.StatusForbidden, gin.H{"error": domainErr.Code, "message": domainErr.Message})
		case domain.ErrRefreshTokenRevoked, domain.ErrRefreshTokenMismatch:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": domainErr.Code, "message": domainErr.Message})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": domainErr.Code, "message": domainErr.Message})
		}
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR", "message": err.Error()})
}

func ExtractBearerToken(header string) string {
	const prefix = "Bearer "
	if header == "" {
		return ""
	}
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}
