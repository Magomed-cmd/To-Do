package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/analytics-service/internal/domain"
	"todoapp/services/analytics-service/internal/ports"
)

type Server struct {
	analyticsv1.UnimplementedAnalyticsServiceServer

	service ports.AnalyticsService
}

func NewServer(service ports.AnalyticsService) *Server {
	return &Server{service: service}
}

func (s *Server) TrackTaskEvent(ctx context.Context, req *analyticsv1.TrackTaskEventRequest) (*analyticsv1.TrackTaskEventResponse, error) {
	input := ports.TrackEventInput{
		Type:       req.GetType(),
		UserID:     req.GetUserId(),
		TaskID:     req.GetTaskId(),
		Status:     req.GetStatus(),
		Priority:   req.GetPriority(),
		OccurredAt: time.Unix(req.GetOccurredAt(), 0),
	}

	if err := s.service.TrackTaskEvent(ctx, input); err != nil {
		return nil, mapError(err)
	}

	return &analyticsv1.TrackTaskEventResponse{}, nil
}

func (s *Server) GetDailyMetrics(ctx context.Context, req *analyticsv1.GetDailyMetricsRequest) (*analyticsv1.GetDailyMetricsResponse, error) {
	var date time.Time
	if req.GetDate() != "" {
		parsed, err := time.Parse("2006-01-02", req.GetDate())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid date format, expected YYYY-MM-DD")
		}
		date = parsed
	}

	metrics, err := s.service.GetDailyMetrics(ctx, ports.DailyMetricsRequest{UserID: req.GetUserId(), Date: date})
	if err != nil {
		return nil, mapError(err)
	}

	return &analyticsv1.GetDailyMetricsResponse{
		Metrics: &analyticsv1.DailyTaskMetrics{
			UserId:         metrics.UserID,
			Date:           metrics.Date.Format("2006-01-02"),
			CreatedTasks:   metrics.CreatedTasks,
			CompletedTasks: metrics.CompletedTasks,
			TotalTasks:     metrics.TotalTasks,
		},
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
