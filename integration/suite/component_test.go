//go:build integration

package suite

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"UserManagement/integration/test_util"
	"UserManagement/internal/util"
)

const (
	restURL      = "http://localhost:8080"
	websocketURL = "ws://localhost:8082/ws_users"
	testTimeout  = 20 * time.Second
)

func setupWebSocket(t *testing.T) *test_util.WebSocketTestUtil {
	wsUtil, err := test_util.NewWebSocketTestUtil(websocketURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	return wsUtil
}

func createUser(t *testing.T) int {
	user := map[string]interface{}{
		"first_name": util.RandomName(),
		"last_name":  util.RandomName(),
		"email":      util.RandomEmail(),
		"phone":      util.RandomPhone(),
		"age":        util.RandomAge(),
		"status":     util.RandomStatus(),
	}
	payload, _ := json.Marshal(user)
	resp, err := http.Post(restURL+"/users", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected HTTP 201 Created")

	var createdUser map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	userID, ok := createdUser["id"].(float64) // JSON numbers are decoded as float64
	if !ok {
		t.Fatalf("Invalid or missing user_id in response: %v", createdUser)
	}
	return int(userID)
}

func TestCreateUserComponent(t *testing.T) {
	// Start WebSocket connection
	wsUtil := setupWebSocket(t)
	defer wsUtil.Close()

	// Prepare REST request payload
	user := map[string]interface{}{
		"first_name": util.RandomName(),
		"last_name":  util.RandomName(),
		"email":      util.RandomEmail(),
		"phone":      util.RandomPhone(),
		"age":        util.RandomAge(),
		"status":     util.RandomStatus(),
	}
	payload, _ := json.Marshal(user)

	// Send REST request
	resp, err := http.Post(restURL+"/users", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected HTTP 201 Created")

	// Listen for WebSocket response
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			msg, ok := wsUtil.GetMessages()
			if !ok {
				continue
			}

			// Validate WebSocket message
			if msg["type"] == "user_list_updated" {
				assert.Equal(t, "success", msg["payload"], "Expected success payload")
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Test timed out waiting for WebSocket response")
	}
}

func TestFetchUsersComponent(t *testing.T) {
	// Start WebSocket connection
	wsUtil := setupWebSocket(t)
	defer wsUtil.Close()

	//Send REST request
	resp, err := http.Get(restURL + "/users")

	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}

	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP 200 OK")
}

func TestUpdateUserComponent(t *testing.T) {
	// Start WebSocket connection
	wsUtil := setupWebSocket(t)
	defer wsUtil.Close()

	// Create User
	userID := createUser(t)

	updateUser := map[string]interface{}{
		"first_name": util.RandomName(),
	}
	updateUserPayload, _ := json.Marshal(updateUser)
	req, err := http.NewRequest(http.MethodPatch, restURL+"/users/"+strconv.Itoa(userID), bytes.NewBuffer(updateUserPayload))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP 200 OK")
}

func TestDeleteUserComponent(t *testing.T) {
	// Start WebSocket connection
	wsUtil := setupWebSocket(t)
	defer wsUtil.Close()

	// Create a user first
	userID := createUser(t)

	// Delete the user
	req, _ := http.NewRequest(http.MethodDelete, restURL+"/users/"+strconv.Itoa(userID), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected HTTP 204 No Content")
}
