package repository

import (
	"context"
	"database/sql"

	sqlc "UserManagement/internal/db/sqlc"
	"UserManagement/internal/model"
	"UserManagement/internal/util"
)

// PostgresUserRepository is the PostgreSQL implementation of UserRepository
type PostgresUserRepository struct {
	queries *sqlc.Queries
}

// NewPostgresUserRepository creates a new instance of PostgresUserRepository
func NewPostgresUserRepository(queries *sqlc.Queries) *PostgresUserRepository {
	return &PostgresUserRepository{queries: queries}
}

func (r *PostgresUserRepository) CreateUserRepo(ctx context.Context, req model.CreateUserRequest) (model.User, error) {
	arg := sqlc.CreateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone: sql.NullString{
			String: req.Phone,
			Valid:  req.Phone != "",
		},
		Age: sql.NullInt32{
			Int32: int32(req.Age),
			Valid: req.Age != 0,
		},
		Status: sql.NullString{
			String: req.Status,
			Valid:  req.Status != "",
		},
	}
	user, err := r.queries.CreateUser(ctx, arg)
	if err != nil {
		return model.User{}, err
	}
	return mapToModelUser(user), nil
}

func (r *PostgresUserRepository) GetUserRepo(ctx context.Context, userID int64) (model.User, error) {
	user, err := r.queries.GetUser(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	return mapToModelUser(user), nil
}

func (r *PostgresUserRepository) UpdateUserRepo(ctx context.Context, userID int64, req model.UpdateUserRequest) (model.User, error) {
	arg := sqlc.UpdateUserParams{
		UserID: userID,
		FirstName: sql.NullString{
			String: util.NullSafeString(req.FirstName),
			Valid:  req.FirstName != nil,
		},
		LastName: sql.NullString{
			String: util.NullSafeString(req.LastName),
			Valid:  req.LastName != nil,
		},
		Email: sql.NullString{
			String: util.NullSafeString(req.Email),
			Valid:  req.Email != nil,
		},
		Phone: sql.NullString{
			String: util.NullSafeString(req.Phone),
			Valid:  req.Phone != nil,
		},
		Age: sql.NullInt32{
			Int32: util.NullSafeInt32(req.Age),
			Valid: req.Age != nil,
		},
		Status: sql.NullString{
			String: util.NullSafeString(req.Status),
			Valid:  req.Status != nil,
		},
	}

	user, err := r.queries.UpdateUser(ctx, arg)
	if err != nil {
		return model.User{}, err
	}
	return mapToModelUser(user), nil
}

func (r *PostgresUserRepository) DeleteUserRepo(ctx context.Context, userID int64) error {
	return r.queries.DeleteUser(ctx, userID)
}

func (r *PostgresUserRepository) ListUsersRepo(ctx context.Context) ([]model.User, error) {
	users, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	var result []model.User
	for _, user := range users {
		result = append(result, mapToModelUser(user))
	}
	return result, nil
}

func mapToModelUser(u sqlc.User) model.User {
	return model.User{
		ID:        u.UserID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     util.NullableStringPtr(u.Phone),
		Age:       util.NullableInt32Ptr(u.Age),
		Status:    util.NullableStringPtr(u.Status),
	}
}
