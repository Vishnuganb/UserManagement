package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ParseAndValidateUserID parses and validates the user ID from the request.
func ParseAndValidateUserID(r *http.Request) (int64, error) {
	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		return 0, errors.New("user ID is required")
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}
	return userID, nil
}

// WriteJSONResponse writes a JSON response to the http.ResponseWriter.
func WriteJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
