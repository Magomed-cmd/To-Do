package analyticsgrpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/task-service/internal/ports"
)

type Config struct {
	Address string
	Timeout time.Duration
}

type grpcConn interface {
	grpc.ClientConnInterface
	Close() error
}

type Client struct {
	conn    grpcConn
	client  analyticsv1.AnalyticsServiceClient
	timeout time.Duration
}

var _ ports.AnalyticsTracker = (*Client)(nil)

var dialGRPC = func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, opts...)
}

func New(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("analyticsgrpc: address is required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Second
	}

	conn, err := dialGRPC(context.Background(), cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		client:  analyticsv1.NewAnalyticsServiceClient(conn),
		timeout: cfg.Timeout,
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) TrackTaskEvent(ctx context.Context, event ports.AnalyticsEvent) error {
	req := &analyticsv1.TrackTaskEventRequest{
		Type:       event.Type,
		UserId:     event.UserID,
		TaskId:     event.TaskID,
		Status:     event.Status,
		Priority:   event.Priority,
		OccurredAt: event.OccurredAt.Unix(),
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if _, err := c.client.TrackTaskEvent(ctx, req); err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.InvalidArgument {
				return nil
			}
		}
		return err
	}

	return nil
}
