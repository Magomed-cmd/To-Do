package metrics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"todoapp/services/analytics-service/internal/domain/entities"
	"todoapp/services/analytics-service/internal/ports"
)

type analyticsStub struct {
	metrics *entities.DailyTaskMetrics
	err     error
	lastReq ports.DailyMetricsRequest
}

func (s *analyticsStub) TrackTaskEvent(ctx context.Context, input ports.TrackEventInput) error {
	return nil
}

func (s *analyticsStub) GetDailyMetrics(ctx context.Context, req ports.DailyMetricsRequest) (*entities.DailyTaskMetrics, error) {
	s.lastReq = req
	return s.metrics, s.err
}

func TestGetDailyMetricsValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := New(&analyticsStub{})
	router := gin.New()
	handler.RegisterRoutes(router)

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{name: "invalid user id", path: "/metrics/daily/not-int", status: http.StatusBadRequest},
		{name: "invalid date", path: "/metrics/daily/1?date=bad-date", status: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rec.Code)
			}
		})
	}
}

func TestGetDailyMetricsServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := New(&analyticsStub{err: errors.New("boom")})
	router := gin.New()
	handler.RegisterRoutes(router)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/daily/3", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestGetDailyMetricsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	date := time.Date(2024, 5, 6, 0, 0, 0, 0, time.UTC)
	metrics := &entities.DailyTaskMetrics{
		UserID:         11,
		Date:           date,
		CreatedTasks:   2,
		CompletedTasks: 1,
		TotalTasks:     3,
		UpdatedAt:      date.Add(time.Hour),
	}
	stub := &analyticsStub{metrics: metrics}

	handler := New(stub)
	router := gin.New()
	handler.RegisterRoutes(router)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/daily/11?date=2024-05-06", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if stub.lastReq.UserID != 11 || stub.lastReq.Date.Year() != 2024 {
		t.Fatalf("unexpected request passed to service: %+v", stub.lastReq)
	}
}
