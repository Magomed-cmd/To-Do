package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"todoapp/pkg/errors"
)

func WriteValidationError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error":   "VALIDATION_ERROR",
		"message": err.Error(),
	})
}

func WriteDomainError(ctx *gin.Context, err error) {
	appErr := errors.AsAppError(err)
	ctx.JSON(appErr.HTTPStatus(), gin.H{
		"error":   appErr.Code,
		"message": appErr.Error(),
	})
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
