package usergrpc

import (
	"context"
	stderrors "errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"todoapp/pkg/errors"
	userv1 "todoapp/pkg/proto/user/v1"
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
	client  userv1.UserServiceClient
	timeout time.Duration
}

var _ ports.UserDirectory = (*Client)(nil)

var dialGRPC = func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, opts...)
}

func New(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, stderrors.New("usergrpc: address is required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}

	conn, err := dialGRPC(context.Background(), cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		client:  userv1.NewUserServiceClient(conn),
		timeout: cfg.Timeout,
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) GetUser(ctx context.Context, userID int64) (*ports.UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &userv1.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, translateError(err)
	}

	user := resp.GetUser()
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	return &ports.UserInfo{
		ID:     user.GetId(),
		Email:  user.GetEmail(),
		Name:   user.GetName(),
		Role:   user.GetRole(),
		Active: user.GetIsActive(),
	}, nil
}

func translateError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return errors.ErrInternal.WithCause(err)
	}

	switch st.Code() {
	case codes.NotFound:
		return errors.ErrUserNotFound
	case codes.PermissionDenied:
		return errors.ErrForbidden
	case codes.InvalidArgument:
		return errors.ErrValidation.WithDetails(st.Message())
	default:
		return errors.ErrInternal.WithCause(err)
	}
}
