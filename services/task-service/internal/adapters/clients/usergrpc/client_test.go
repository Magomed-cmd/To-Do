package usergrpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "todoapp/pkg/proto/user/v1"
	"todoapp/services/task-service/internal/domain"
)

type fakeUserConn struct{}

func (fakeUserConn) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	return nil
}
func (fakeUserConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}
func (fakeUserConn) Close() error { return nil }

type userClientStub struct {
	err      error
	response *userv1.GetUserResponse
	req      *userv1.GetUserRequest
}

func (s *userClientStub) GetUser(ctx context.Context, in *userv1.GetUserRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	s.req = in
	return s.response, s.err
}

func (s *userClientStub) ValidateToken(ctx context.Context, in *userv1.ValidateTokenRequest, opts ...grpc.CallOption) (*userv1.ValidateTokenResponse, error) {
	return nil, nil
}

func TestUserClientNewValidation(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected error on empty address")
	}
}

func TestUserClientNewDefaultTimeout(t *testing.T) {
	origDial := dialGRPC
	defer func() { dialGRPC = origDial }()
	dialGRPC = func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return &grpc.ClientConn{}, nil
	}

	client, err := New(Config{Address: "localhost:0", Timeout: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.timeout != 3*time.Second {
		t.Fatalf("expected default timeout, got %v", client.timeout)
	}
}

func TestGetUserSuccess(t *testing.T) {
	stub := &userClientStub{
		response: &userv1.GetUserResponse{
			User: &userv1.User{
				Id:       10,
				Email:    "user@example.com",
				Name:     "User",
				Role:     "admin",
				IsActive: true,
			},
		},
	}
	client := &Client{client: stub, timeout: time.Second}

	user, err := client.GetUser(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 10 || user.Email != "user@example.com" || user.Role != "admin" || !user.Active {
		t.Fatalf("unexpected user: %+v", user)
	}
	if stub.req == nil || stub.req.UserId != 10 {
		t.Fatalf("request not sent to stub")
	}
}

func TestGetUserErrorTranslation(t *testing.T) {
	tests := []struct {
		err        error
		expectCode string
		expectMsg  string
	}{
		{status.Error(codes.NotFound, ""), domain.ErrUnknownUser.Code, domain.ErrUnknownUser.Message},
		{status.Error(codes.PermissionDenied, ""), domain.ErrForbiddenTaskAccess.Code, domain.ErrForbiddenTaskAccess.Message},
		{status.Error(codes.InvalidArgument, "bad"), domain.ErrValidationFailed.Code, "bad"},
		{errors.New("other"), "", "other"},
		{nil, domain.ErrUnknownUser.Code, domain.ErrUnknownUser.Message}, // nil user in response
	}

	for _, tt := range tests {
		stub := &userClientStub{err: tt.err, response: &userv1.GetUserResponse{}}
		client := &Client{client: stub, timeout: time.Second}

		if tt.err == nil {
			stub.response = &userv1.GetUserResponse{User: nil}
		}

		_, err := client.GetUser(context.Background(), 1)

		if tt.expectCode == "" {
			if err == nil || err.Error() != tt.expectMsg {
				t.Fatalf("expected raw error %q, got %v", tt.expectMsg, err)
			}
			continue
		}

		var domainErr *domain.DomainError
		if !errors.As(err, &domainErr) {
			t.Fatalf("expected domain error, got %v", err)
		}
		if domainErr.Code != tt.expectCode || domainErr.Message != tt.expectMsg {
			t.Fatalf("unexpected domain error: %+v", domainErr)
		}
	}
}
