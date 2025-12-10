package ports

import "context"

// UserInfo represents a lightweight user profile returned by the user-service via gRPC.
type UserInfo struct {
	ID     int64
	Email  string
	Name   string
	Role   string
	Active bool
}

// UserDirectory exposes the operations required from the user-service.
type UserDirectory interface {
	GetUser(ctx context.Context, userID int64) (*UserInfo, error)
}
