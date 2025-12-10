package analyticsgrpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/task-service/internal/ports"
)

type fakeConn struct{}

func (fakeConn) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	return nil
}
func (fakeConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}
func (fakeConn) Close() error { return nil }

type analyticsClientStub struct {
	err      error
	lastReq  *analyticsv1.TrackTaskEventRequest
	metrics  *analyticsv1.GetDailyMetricsResponse
	metricsE error
}

func (c *analyticsClientStub) TrackTaskEvent(ctx context.Context, in *analyticsv1.TrackTaskEventRequest, opts ...grpc.CallOption) (*analyticsv1.TrackTaskEventResponse, error) {
	c.lastReq = in
	return &analyticsv1.TrackTaskEventResponse{}, c.err
}

func (c *analyticsClientStub) GetDailyMetrics(ctx context.Context, in *analyticsv1.GetDailyMetricsRequest, opts ...grpc.CallOption) (*analyticsv1.GetDailyMetricsResponse, error) {
	return c.metrics, c.metricsE
}

func TestNewValidation(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected error on empty address")
	}
}

func TestNewSetsDefaultTimeout(t *testing.T) {
	origDial := dialGRPC
	defer func() { dialGRPC = origDial }()
	dialGRPC = func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return &grpc.ClientConn{}, nil
	}

	client, err := New(Config{Address: "localhost:0", Timeout: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.timeout != 2*time.Second {
		t.Fatalf("expected default timeout to be set, got %v", client.timeout)
	}
}

func TestTrackTaskEvent(t *testing.T) {
	tests := []struct {
		name      string
		stubErr   error
		wantError bool
	}{
		{name: "success", stubErr: nil, wantError: false},
		{name: "invalid argument ignored", stubErr: status.Error(codes.InvalidArgument, "bad"), wantError: false},
		{name: "other error returned", stubErr: errors.New("boom"), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &analyticsClientStub{err: tt.stubErr}
			client := &Client{client: stub, timeout: time.Second}

			err := client.TrackTaskEvent(context.Background(), ports.AnalyticsEvent{
				Type:     analyticsv1.TaskEventType_TASK_EVENT_TYPE_COMPLETED,
				UserID:   1,
				TaskID:   2,
				Status:   "completed",
				Priority: "high",
			})

			if tt.wantError && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stub.lastReq == nil || stub.lastReq.TaskId != 2 || stub.lastReq.UserId != 1 {
				t.Fatalf("request was not forwarded")
			}
		})
	}
}
