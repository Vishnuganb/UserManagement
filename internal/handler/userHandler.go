package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"

	"UserManagement/internal/model"
)

type UserService interface {
	QueueCUDRequest(req model.CUDRequest)
}

type UserHandler struct {
	us UserService
}

func NewUserHandler(us UserService) *UserHandler {
	return &UserHandler{us: us}
}

// Common function to handle requests
func (h *UserHandler) handleRequest(w http.ResponseWriter, cudReq model.CUDRequest, successStatus int) {
	responseChan := make(chan interface{})
	cudReq.ResponseChannel = responseChan
	h.us.QueueCUDRequest(cudReq)

	select {
	case response := <-responseChan:
		if err, ok := response.(error); ok {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "User Not Found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		w.WriteHeader(successStatus)
		writeJSON(w, response)
	case <-time.After(5 * time.Second): // Timeout after 5 seconds
		http.Error(w, "Request timed out", http.StatusGatewayTimeout)
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send the request to the channel for asynchronous processing
	cudReq := model.CUDRequest{
		Type:      "create_user",
		CreateReq: req,
	}
	h.handleRequest(w, cudReq, http.StatusCreated)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, _ *http.Request) {
	cudReq := model.CUDRequest{
		Type: "get_users",
	}
	h.handleRequest(w, cudReq, http.StatusOK)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	// Parse string to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	cudReq := model.CUDRequest{
		Type:   "get_user",
		UserID: userID,
	}
	h.handleRequest(w, cudReq, http.StatusOK)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	// Parse string to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	cudReq := model.CUDRequest{
		Type:   "delete_user",
		UserID: userID,
	}
	h.handleRequest(w, cudReq, http.StatusNoContent)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	// Parse string to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cudReq := model.CUDRequest{
		Type: "update_user",
		UpdateReq: struct {
			UserID int64
			Req    model.UpdateUserRequest
		}{
			UserID: userID,
			Req:    req,
		},
	}
	h.handleRequest(w, cudReq, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	switch v := data.(type) {
	case error:
		w.WriteHeader(http.StatusBadRequest) // Use 400 for errors
		response := map[string]string{"error": v.Error()}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Failed to encode error response:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusOK) // Use 200 for successful responses
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Println("Failed to encode response:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
