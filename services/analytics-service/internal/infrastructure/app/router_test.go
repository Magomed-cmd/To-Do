package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"todoapp/services/analytics-service/internal/domain/entities"
	"todoapp/services/analytics-service/internal/ports"
)

type analyticsMock struct{}

func (analyticsMock) TrackTaskEvent(ctx context.Context, input ports.TrackEventInput) error { return nil }
func (analyticsMock) GetDailyMetrics(ctx context.Context, req ports.DailyMetricsRequest) (*entities.DailyTaskMetrics, error) {
	return &entities.DailyTaskMetrics{UserID: req.UserID, Date: req.Date}, nil
}

func TestNewRouterValidation(t *testing.T) {
	_, err := NewRouter(HTTPDeps{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := NewRouter(HTTPDeps{Analytics: analyticsMock{}, ServiceName: "analytics"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name   string
		method string
		status int
	}{
		{name: "get", method: http.MethodGet, status: http.StatusOK},
		{name: "head", method: http.MethodHead, status: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/health", nil)
			router.ServeHTTP(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rec.Code)
			}
		})
	}
}
