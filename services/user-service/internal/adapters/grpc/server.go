package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "todoapp/pkg/proto/user/v1"
	"todoapp/services/user-service/internal/domain"
	"todoapp/services/user-service/internal/domain/entities"
	"todoapp/services/user-service/internal/ports"
)

// Server exposes the user service capabilities over gRPC for other services.
type Server struct {
	userv1.UnimplementedUserServiceServer

	service ports.UserService
	tokens  ports.TokenManager
}

func NewServer(service ports.UserService, tokens ports.TokenManager) *Server {
	return &Server{service: service, tokens: tokens}
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.service.GetProfile(ctx, req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &userv1.GetUserResponse{User: toProtoUser(*user)}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *userv1.ValidateTokenRequest) (*userv1.ValidateTokenResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "access token is required")
	}

	claims, err := s.tokens.ParseAccessToken(req.GetAccessToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid access token")
	}

	return &userv1.ValidateTokenResponse{
		UserId:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

func toProtoUser(user entities.User) *userv1.User {
	return &userv1.User{
		Id:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
		IsActive: user.IsActive,
	}
}

func mapDomainError(err error) error {
	if err == nil {
		return status.Error(codes.Internal, "unknown error")
	}

	if domainErr, ok := err.(*domain.DomainError); ok {
		switch domainErr {
		case domain.ErrUserNotFound:
			return status.Error(codes.NotFound, domainErr.Message)
		case domain.ErrUserInactive, domain.ErrUserSuspended, domain.ErrUserLocked:
			return status.Error(codes.PermissionDenied, domainErr.Message)
		default:
			return status.Error(codes.InvalidArgument, domainErr.Message)
		}
	}

	return status.Error(codes.Internal, err.Error())
}
