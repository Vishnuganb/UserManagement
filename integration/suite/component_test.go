//go:build integration

package suite

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"

	"UserManagement/integration/test_util"
	"UserManagement/internal/util"
)

// --- Tests ----

func setupWebSocket(t *testing.T) *test_util.WebSocketTestUtil {
	// Start WebSocket connection
	wsUtil := test_util.SetupWebSocket(t)
	t.Cleanup(func() {
		wsUtil.Close()
	})
	return wsUtil
}

func TestCreateUserComponent(t *testing.T) {
	wsUtil := setupWebSocket(t)

	// Prepare REST request payload
	payload, _ := json.Marshal(test_util.CreateUserPayload())

	// Send REST request
	resp, err := http.Post(test_util.RestURL+"/users", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected HTTP 201 Created")

	// Validate response keys dynamically
	test_util.ValidateResponseKeys(t, resp)

	// Listen for WebSocket response
	test_util.WaitForWebSocketEvent(t, wsUtil)
}

func TestFetchUsersComponent(t *testing.T) {
	//Send REST request
	resp, err := http.Get(test_util.RestURL + "/users")
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}

	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP 200 OK")

	// Validate response keys dynamically
	test_util.ValidateResponseKeys(t, resp)
}

func TestUpdateUserComponent(t *testing.T) {
	wsUtil := setupWebSocket(t)

	// Create User
	userID := test_util.CreateUser(t)

	updateUser := map[string]interface{}{
		"first_name": util.RandomName(),
	}
	updateUserPayload, _ := json.Marshal(updateUser)
	req, err := http.NewRequest(http.MethodPatch, test_util.RestURL+"/users/"+strconv.Itoa(userID), bytes.NewBuffer(updateUserPayload))
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

	// Validate response keys dynamically
	test_util.ValidateResponseKeys(t, resp)

	// Listen for WebSocket response
	test_util.WaitForWebSocketEvent(t, wsUtil)
}

func TestDeleteUserComponent(t *testing.T) {
	wsUtil := setupWebSocket(t)

	// Create a user first
	userID := test_util.CreateUser(t)

	// Delete the user
	req, _ := http.NewRequest(http.MethodDelete, test_util.RestURL+"/users/"+strconv.Itoa(userID), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send REST request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP 200 OK")

	// Validate response keys dynamically
	test_util.ValidateResponseKeys(t, resp)

	// Listen for WebSocket response
	test_util.WaitForWebSocketEvent(t, wsUtil)
}
