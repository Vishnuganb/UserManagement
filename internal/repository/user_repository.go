package repository

import (
	"context"

	"UserManagement/internal/model"
)

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	CreateUserRepo(ctx context.Context, req model.CreateUserRequest) (model.User, error)
	GetUserRepo(ctx context.Context, userID int64) (model.User, error)
	UpdateUserRepo(ctx context.Context, userID int64, req model.UpdateUserRequest) (model.User, error)
	DeleteUserRepo(ctx context.Context, userID int64) (model.User, error)
	ListUsersRepo(ctx context.Context) ([]model.User, error)
}
