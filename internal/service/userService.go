package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"UserManagement/internal/model"
	"UserManagement/internal/repository"
)

type UserNotifier interface {
	NotifyUserCreated(key, value string) error
}

type Validator interface {
	ValidateCreateUser(firstName, lastName, email string) error
}

type UserService struct {
	repo     repository.UserRepository
	v        Validator
	notifier UserNotifier
	channel  chan model.CUDRequest
}

func NewUserService(ctx context.Context, repo repository.UserRepository, v Validator, notifier UserNotifier) *UserService {
	us := &UserService{
		repo:     repo,
		v:        v,
		notifier: notifier,
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
				if user, err := s.CreateUser(ctx, req.CreateReq); err != nil {
					log.Printf("Error processing user creation from channel: %v\n", err)
					req.ResponseChannel <- err
				} else {
					req.ResponseChannel <- user
				}
			case "update_user":
				log.Printf("Processing user update from channel: %+v\n", req.UpdateReq)
				if user, err := s.UpdateUser(ctx, req.UpdateReq.UserID, req.UpdateReq.Req); err != nil {
					log.Printf("Error processing user update: %v\n", err)
					req.ResponseChannel <- err
				} else {
					req.ResponseChannel <- user
				}
			case "delete_user":
				log.Printf("Processing user deletion from channel: %+v\n", req.UserID)
				if user, err := s.DeleteUser(ctx, req.UserID); err != nil {
					log.Printf("Error processing user deletion: %v\n", err)
					req.ResponseChannel <- err
				} else {
					req.ResponseChannel <- user
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

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error) {
	if err := s.v.ValidateCreateUser(req.FirstName, req.LastName, req.Email); err != nil {
		log.Println("Validation failed:", err)
		return model.User{}, err
	}

	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.repo.CreateUserRepo(ctx, req)
	if err != nil {
		log.Println("Failed to create user:", err)
		return model.User{}, err
	}

	// Publish a message to Kafka
	message := fmt.Sprintf("User %s %s created", req.FirstName, req.LastName)
	if err := s.notifier.NotifyUserCreated("user_created", message); err != nil {
		log.Println("Failed to publish Kafka message:", err)
	}

	return user, nil
}

func (s *UserService) GetUsers(ctx context.Context) ([]model.User, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	users, err := s.repo.ListUsersRepo(ctx)
	return users, err
}

func (s *UserService) GetUserById(ctx context.Context, userId int64) (model.User, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.repo.GetUserRepo(ctx, userId)
	return user, err
}

func (s *UserService) DeleteUser(ctx context.Context, userId int64) (model.User, error) {
	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.repo.DeleteUserRepo(ctx, userId)
	// Publish a message to Kafka
	message := fmt.Sprintf("User %s deleted", user)
	if err := s.notifier.NotifyUserCreated("user_deleted", message); err != nil {
		log.Println("Failed to publish Kafka message:", err)
	}
	return user, err
}

func (s *UserService) UpdateUser(ctx context.Context, userId int64, req model.UpdateUserRequest) (model.User, error) {

	// Create a new context with a deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.repo.UpdateUserRepo(ctx, userId, req)
	// Publish a message to Kafka
	message := fmt.Sprintf("User %s created", user)
	if err := s.notifier.NotifyUserCreated("user_updated", message); err != nil {
		log.Println("Failed to publish Kafka message:", err)
	}
	return user, err
}
