package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/ports"
)

const ContextUserClaimsKey = "taskService.userClaims"

type Middleware struct {
	tokens ports.TokenManager
}

func New(tokens ports.TokenManager) *Middleware {
	return &Middleware{tokens: tokens}
}

func (m *Middleware) JWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "MISSING_TOKEN"})
			return
		}

		token := extractBearerToken(header)
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "INVALID_TOKEN"})
			return
		}

		claims, err := m.tokens.ParseAccessToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "INVALID_TOKEN"})
			return
		}

		ctx.Set(ContextUserClaimsKey, claims)
		ctx.Next()
	}
}

func CurrentUser(ctx *gin.Context) (*ports.TokenClaims, bool) {
	value, ok := ctx.Get(ContextUserClaimsKey)
	if !ok {
		return nil, false
	}

	claims, ok := value.(*ports.TokenClaims)
	return claims, ok
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(header[len(prefix):])
}
