package handler

import (
	"UserManagement/internal/errs"
	"UserManagement/internal/model"
	"UserManagement/internal/util"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
func (h *UserHandler) handleRequest(ctx context.Context, w http.ResponseWriter, cudReq model.CUDRequest, successStatus int) {
	responseChan := make(chan interface{})
	cudReq.ResponseChannel = responseChan
	h.us.QueueCUDRequest(cudReq)

	select {
	case response := <-responseChan:
		if err, ok := response.(error); ok {
			switch {
			case errors.Is(err, errs.ErrUserNotFound):
				http.Error(w, "User Not Found", http.StatusNotFound)
			case errors.Is(err, errs.ErrDuplicateUser):
				http.Error(w, "User Already Exists", http.StatusBadRequest)
			default:
				log.Println("Unhandled error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(successStatus)
		writeJSON(w, response)
	case <-ctx.Done():
		http.Error(w, "Request timed out", http.StatusGatewayTimeout)
	}
}

func (h *UserHandler) parseAndValidateUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := util.ParseAndValidateUserID(r)
	if err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, util.APIResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return 0, false
	}
	return userID, true
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	cudReq := model.CUDRequest{
		Type:      "create_user",
		CreateReq: req,
	}
	h.handleRequest(r.Context(), w, cudReq, http.StatusCreated)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	cudReq := model.CUDRequest{
		Type: "get_users",
	}
	h.handleRequest(r.Context(), w, cudReq, http.StatusOK)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.parseAndValidateUserID(w, r)
	if !ok {
		return
	}
	cudReq := model.CUDRequest{
		Type:   "get_user",
		UserID: userID,
	}
	h.handleRequest(r.Context(), w, cudReq, http.StatusOK)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.parseAndValidateUserID(w, r)
	if !ok {
		return
	}
	cudReq := model.CUDRequest{
		Type:   "delete_user",
		UserID: userID,
	}
	h.handleRequest(r.Context(), w, cudReq, http.StatusOK)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.parseAndValidateUserID(w, r)
	if !ok {
		return
	}
	var req model.UpdateUserRequest
	if !decodeJSON(w, r, &req) {
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
	h.handleRequest(r.Context(), w, cudReq, http.StatusOK)
}

// Helper to decode JSON with error handling
func decodeJSON(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := data.(error); ok {
		w.WriteHeader(http.StatusBadRequest) // Use 400 for errors
	} else {
		w.WriteHeader(http.StatusOK) // Use 200 for successful responses
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Failed to encode response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
