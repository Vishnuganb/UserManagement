package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

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

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send the request to the channel for asynchronous processing
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "create_user",
		CreateReq:       req,
		ResponseChannel: responseChan,
	}
	h.us.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, _ *http.Request) {
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "get_users",
		ResponseChannel: responseChan,
	}
	h.us.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	writeJSON(w, response)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	// Parse string to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "get_user",
		UserID:          userID,
		ResponseChannel: responseChan,
	}
	h.us.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	writeJSON(w, response)
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

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	// Parse string to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "delete_user",
		UserID:          userID,
		ResponseChannel: responseChan,
	}
	h.us.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	w.WriteHeader(http.StatusOK)
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
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type: "update_user",
		UpdateReq: struct {
			UserID int64
			Req    model.UpdateUserRequest
		}{
			UserID: userID,
			Req:    req,
		},
		ResponseChannel: responseChan,
	}
	h.us.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	w.WriteHeader(http.StatusOK)
}
