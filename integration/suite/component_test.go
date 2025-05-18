//go:build integration

package suite

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"UserManagement/internal/util"
)

const (
	restURL      = "http://localhost:8080"
	websocketURL = "ws://localhost:8082/ws_users"
	testTimeout  = 20 * time.Second
)

func TestCreateUserComponent(t *testing.T) {
	// Start WebSocket connection
	wsConn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsConn.Close()

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
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
			log.Printf("WebSocket message received: %s", message)

			// Validate WebSocket message
			var response map[string]interface{}
			_ = json.Unmarshal(message, &response)
			if response["type"] == "user_list_updated" {
				assert.Equal(t, "success", response["payload"], "Expected success payload")
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
	wsConn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsConn.Close()

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
	wsConn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsConn.Close()

	// Create User
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

	// Extract User ID from response
	var createdUser string
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	updateUser := map[string]interface{}{
		"first_name": util.RandomName(),
	}
	updateUserPayload, _ := json.Marshal(updateUser)
	req, err := http.NewRequest(http.MethodPatch, restURL+"/users/2", bytes.NewBuffer(updateUserPayload))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP 200 OK")
}

func TestDeleteUserComponent(t *testing.T) {
	// Start WebSocket connection
	wsConn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsConn.Close()

	// Create a user first
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

	// Delete the user
	req, _ := http.NewRequest(http.MethodDelete, restURL+"/users/1", nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected HTTP 204 No Content")
}
