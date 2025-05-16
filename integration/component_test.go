package integration

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

const (
	restURL      = "http://localhost:8080"
	websocketURL = "ws://localhost:8082/ws_users"
	testTimeout  = 10 * time.Second
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
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"phone":      "1234567890",
		"age":        30,
		"status":     "active",
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
