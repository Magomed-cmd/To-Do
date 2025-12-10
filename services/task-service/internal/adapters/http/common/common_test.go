package common

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"todoapp/services/task-service/internal/domain"
)

func TestWriteValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	WriteValidationError(ctx, errors.New("bad"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWriteDomainError(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{domain.ErrTaskNotFound, http.StatusNotFound},
		{domain.ErrForbiddenTaskAccess, http.StatusForbidden},
		{domain.ErrInvalidTaskPriority, http.StatusBadRequest},
		{errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		WriteDomainError(ctx, tt.err)
		if rec.Code != tt.status {
			t.Fatalf("expected %d for %v, got %d", tt.status, tt.err, rec.Code)
		}
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		expect string
	}{
		{name: "empty", header: "", expect: ""},
		{name: "no prefix", header: "token", expect: ""},
		{name: "with prefix", header: "Bearer abc", expect: "abc"},
	}

	for _, tt := range tests {
		if got := ExtractBearerToken(tt.header); got != tt.expect {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.expect, got)
		}
	}
}
