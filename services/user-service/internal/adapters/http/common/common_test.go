package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"todoapp/pkg/errors"
)

func TestExtractBearerToken(t *testing.T) {
	token := ExtractBearerToken("Bearer abc")
	if token != "abc" {
		t.Fatalf("expected token abc, got %s", token)
	}
	if ExtractBearerToken("abc") != "" {
		t.Fatalf("expected empty token for missing prefix")
	}
}

func TestWriteDomainError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	WriteDomainError(ctx, errors.ErrInvalidCredentials)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if body["error"] != string(errors.CodeInvalidCredentials) {
		t.Fatalf("unexpected error code %s", body["error"])
	}
}

func TestWriteValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	WriteValidationError(ctx, errValidation{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

type errValidation struct{}

func (errValidation) Error() string {
	return "validation"
}
