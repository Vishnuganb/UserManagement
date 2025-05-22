package test_util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"UserManagement/internal/util"
	"github.com/stretchr/testify/assert"
)

const (
	RestURL      = "http://localhost:8080"
	WebSocketURL = "ws://localhost:8082/ws_users"
	TestTimeout  = 20 * time.Second
)

func SetupWebSocket(t *testing.T) *WebSocketTestUtil {
	wsUtil, err := NewWebSocketTestUtil(WebSocketURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	return wsUtil
}

func CreateUserPayload() map[string]interface{} {
	return map[string]interface{}{
		"first_name": util.RandomName(),
		"last_name":  util.RandomName(),
		"email":      util.RandomEmail(),
		"phone":      util.RandomPhone(),
		"age":        util.RandomAge(),
		"status":     util.RandomStatus(),
	}
}

func CreateUser(t *testing.T) int {
	user := CreateUserPayload()
	payload, _ := json.Marshal(user)
	resp, err := http.Post(RestURL+"/users", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected HTTP 201 Created")

	var createdUser map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	userID, ok := createdUser["id"].(float64)
	if !ok {
		t.Fatalf("Invalid or missing user_id in response: %v", createdUser)
	}
	return int(userID)
}

func WaitForWebSocketEvent(t *testing.T, wsUtil *WebSocketTestUtil) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			msg, ok := wsUtil.GetMessages()
			if !ok {
				continue
			}
			switch msg["type"] {
			case "New User Created", "User Deleted", "User Updated":
				assert.Equal(t, "success", msg["payload"], "Expected success payload")
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(TestTimeout):
		t.Fatal("Test timed out waiting for WebSocket response")
	}
}

// ValidateResponseKeys validates the response keys dynamically
func ValidateResponseKeys(t *testing.T, response *http.Response) {
	var data interface{}
	err := json.NewDecoder(response.Body).Decode(&data)
	assert.NoError(t, err, "Failed to decode response payload")

	switch v := data.(type) {
	case []interface{}: // Handle array of users
		for _, item := range v {
			user, ok := item.(map[string]interface{})
			assert.True(t, ok, "Expected user to be a map")
			for _, key := range GetExpectedUserKeys() {
				assert.Contains(t, user, key, "Response payload should contain '"+key+"'")
			}
		}
	case map[string]interface{}: // Handle single user object
		for _, key := range GetExpectedUserKeys() {
			assert.Contains(t, v, key, "Response payload should contain '"+key+"'")
		}
	default:
		t.Fatalf("Unexpected response format: %T", v)
	}
}

func GetExpectedUserKeys() []string {
	return []string{
		"id",
		"first_name",
		"last_name",
		"email",
		"phone",
		"age",
		"status",
		"created_at",
		"updated_at",
	}
}
