package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/domain"
)

func WriteValidationError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error":   "VALIDATION_ERROR",
		"message": err.Error(),
	})
}

func WriteDomainError(ctx *gin.Context, err error) {
	if domainErr, ok := err.(*domain.DomainError); ok {
		switch domainErr {
		case domain.ErrTaskNotFound, domain.ErrCategoryNotFound, domain.ErrCommentNotFound, domain.ErrUnknownUser:
			ctx.JSON(http.StatusNotFound, gin.H{"error": domainErr.Code, "message": domainErr.Message})
			return
		case domain.ErrForbiddenTaskAccess:
			ctx.JSON(http.StatusForbidden, gin.H{"error": domainErr.Code, "message": domainErr.Message})
			return
		case domain.ErrInvalidTaskPriority, domain.ErrInvalidTaskStatus, domain.ErrValidationFailed, domain.ErrInvalidRecurringRule:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": domainErr.Code, "message": domainErr.Message})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": domainErr.Code, "message": domainErr.Message})
			return
		}
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
