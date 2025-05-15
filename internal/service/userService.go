package service

import (
	sqlc "UserManagement/internal/db/sqlc"
	"UserManagement/internal/model"
	"UserManagement/internal/util"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
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
	channel  chan model.CUDRequest
}

func NewUserService(ctx context.Context, db *sql.DB, v Validator, producer MessageProducer) *UserService {
	us := &UserService{
		db:       db,
		v:        v,
		q:        sqlc.New(db),
		producer: producer,
		channel:  make(chan model.CUDRequest, 100), // Buffered channel to handle multiple requests
	}

	// Start a goroutine to listen for messages on the channel
	go us.listenToChannel(ctx)

	return us
}

func (s *UserService) listenToChannel(ctx context.Context) {
	for {
		select {
		case req := <-s.channel:
			switch req.Type {
			case "create_user":
				log.Printf("Processing user creation from channel: %+v\n", req.CreateReq)
				if err := s.CreateUser(ctx, req.CreateReq); err != nil {
					log.Printf("Error processing user creation from channel: %v\n", err)
					req.ResponseChannel <- err
				}
			case "update_user":
				log.Printf("Processing user update from channel: %+v\n", req.UpdateReq)
				if _, err := s.UpdateUser(ctx, req.UpdateReq.UserID, req.UpdateReq.Req); err != nil {
					log.Printf("Error processing user update: %v\n", err)
					req.ResponseChannel <- err
				}
			case "delete_user":
				log.Printf("Processing user deletion from channel: %+v\n", req.UserID)
				if err := s.DeleteUser(ctx, req.UserID); err != nil {
					log.Printf("Error processing user deletion: %v\n", err)
					req.ResponseChannel <- err
				}
			case "get_users":
				log.Printf("Processing get users request from channel")
				users, err := s.GetUsers(ctx)
				if err != nil {
					log.Printf("Error processing get users request: %v\n", err)
					req.ResponseChannel <- err
				} else {
					req.ResponseChannel <- users
				}
			case "get_user":
				log.Printf("Processing get user request from channel: %+v\n", req.UserID)
				user, err := s.GetUserById(ctx, req.UserID)
				if err != nil {
					log.Printf("Error processing get user request: %v\n", err)
					req.ResponseChannel <- err
				} else {
					req.ResponseChannel <- user
				}
			}

		case <-ctx.Done():
			log.Println("Context canceled, stopping channel listener")
			s.channel = nil // Set channel to nil to indicate no listener
			return
		}
	}
}

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) error {
	if err := s.v.ValidateCreateUser(req.FirstName, req.LastName, req.Email); err != nil {
		return err
	}

	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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

// QueueCUDRequest SendToChannel Add a method to send messages to the channel
func (s *UserService) QueueCUDRequest(req model.CUDRequest) {
	if s.channel == nil {
		log.Println("No active listener on the channel, dropping request")
		return
	}
	select {
	case s.channel <- req:
		log.Println("Request queued successfully")
	case <-time.After(3 * time.Second):
		log.Println("Timeout: failed to queue request")
	}
}

func (s *UserService) GetUsers(ctx context.Context) ([]sqlc.User, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	users, err := s.q.ListUsers(ctx)
	return users, err
}

func (s *UserService) GetUserById(ctx context.Context, userId int64) (sqlc.User, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.q.GetUser(ctx, userId)
	return user, err
}

func (s *UserService) DeleteUser(ctx context.Context, userId int64) error {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.q.DeleteUser(ctx, userId)
}

func (s *UserService) UpdateUser(ctx context.Context, userId int64, req model.UpdateUserRequest) (sqlc.User, error) {
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

	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.q.UpdateUser(ctx, arg)
	return user, err
}
