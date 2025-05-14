package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	sqlc "UserManagement/internal/db/sqlc"
	"UserManagement/internal/model"
	"UserManagement/internal/util"
)

type MessageProducer interface {
	Publish(key, value string) error
}

type Validator interface {
	ValidateCreateUser(firstName, lastName, email string) error
}

type UserService struct {
	db       *sql.DB
	v        Validator
	q        *sqlc.Queries
	producer MessageProducer
}

func NewUserService(db *sql.DB, v Validator, producer MessageProducer) *UserService {
	return &UserService{
		db:       db,
		v:        v,
		q:        sqlc.New(db),
		producer: producer,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) error {
	if err := s.v.ValidateCreateUser(req.FirstName, req.LastName, req.Email); err != nil {
		return err
	}
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
	/*
		query := `INSERT INTO users (first_name, last_name, email, phone, age, status)
		VALUES ($1, $2, $3, $4, $5, $6)`

		_, err := s.db.ExecContext(ctx, query, req.FirstName, req.LastName, req.Email, req.Phone, req.Age, req.Status)

	*/
	_, err := s.q.CreateUser(ctx, arg)
	if err != nil {
		return err
	}

	// Publish a message to Kafka
	message := fmt.Sprintf("User %s %s created", req.FirstName, req.LastName)
	if err := s.producer.Publish("user_created", message); err != nil {
		log.Println("Failed to publish Kafka message:", err)
	}

	return nil
}

func (s *UserService) GetUsers(ctx context.Context) ([]sqlc.User, error) {
	/*
		query := `SELECT * FROM users`
			rows, err := s.db.QueryContext(ctx, query)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var users []model.User

			for rows.Next() {
				var user model.User
				err := rows.Scan( //  Save Current Row into Go variables
					&user.UserID,
					&user.FirstName,
					&user.LastName,
					&user.Email,
					&user.Phone,
					&user.Age,
					&user.Status,
					&user.CreatedAt,
					&user.UpdatedAt)
				if err != nil {
					return nil, err
				}
				users = append(users, user)
			}

			if err := rows.Err(); err != nil {
				return nil, err
			}

	*/
	users, err := s.q.ListUsers(ctx)
	return users, err
}

func (s *UserService) GetUserById(ctx context.Context, userId int64) (sqlc.User, error) {
	/*

		query := `SELECT * FROM users WHERE user_id = $1`
		row := s.db.QueryRowContext(ctx, query, userId)
		var user model.User
		err := row.Scan(
			&user.UserID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Phone,
			&user.Age,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt)

		if err != nil {
			return model.User{}, err
		}

	*/

	user, err := s.q.GetUser(ctx, userId)

	return user, err
}

func (s *UserService) DeleteUser(ctx context.Context, userId int64) error {
	/*
		query := `DELETE FROM users WHERE user_id=$1`
		_, err := s.db.ExecContext(ctx, query, userId)
		if err != nil {
			return err
		}
		return nil
	*/
	return s.q.DeleteUser(ctx, userId)
}

func (s *UserService) UpdateUser(ctx context.Context, userId int64, req model.UpdateUserRequest) (sqlc.User, error) {
	/*
		query := `
			UPDATE users
			SET
				first_name = COALESCE(NULLIF($1, ''), first_name),
				last_name = COALESCE(NULLIF($2, ''), last_name),
				email = COALESCE(NULLIF($3, ''), email),
				phone = COALESCE(NULLIF($4, ''), phone),
				age = COALESCE($5, age),
				status = COALESCE(NULLIF($6, ''), status),
				updated_at = NOW()
			WHERE user_id = $7
			`

		_, err := s.db.ExecContext(ctx, query,
			util.NullSafeString(req.FirstName),
			util.NullSafeString(req.LastName),
			util.NullSafeString(req.Email),
			util.NullSafeString(req.Phone),
			util.NullSafeInt32(req.Age),
			util.NullSafeString(req.Status),
			userId,
		)
	*/
	arg := sqlc.UpdateUserParams{
		UserID: userId,
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

	user, err := s.q.UpdateUser(ctx, arg)
	return user, err
}
