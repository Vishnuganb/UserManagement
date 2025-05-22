package ws

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"UserManagement/internal/model"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type UserService interface {
	QueueCUDRequest(req model.CUDRequest)
}

type Manager struct {
	UserService UserService
	clients     ClientList
	sync.RWMutex
	handlers map[string]MessageHandler
}

func NewManager(us UserService) *Manager {
	m := &Manager{
		UserService: us,
		clients:     make(ClientList),
		handlers:    make(map[string]MessageHandler),
	}
	m.setupMessageHandlers()
	return m
}

func (m *Manager) setupMessageHandlers() {
	m.handlers["create_user"] = m.handleCreateUser
	m.handlers["get_users"] = m.handleGetUsers
	m.handlers["update_user"] = m.handleUpdateUser
	m.handlers["delete_user"] = m.handleDeleteUser
}

func (m *Manager) routeEvent(message Message, c *Client) error {
	// check event type is part of the handler
	if handler, ok := m.handlers[message.Type]; ok {
		if err := handler(message, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("event handler not found")
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new WS Connection")
	// upgrade regular http connection into websocket
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Couldn't able to upgrade", err)
		return
	}
	client := NewClient(conn, m)
	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok {
		err := client.conn.Close()
		if err != nil {
			return
		}
		delete(m.clients, client)
	}
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Allow empty origin (e.g., Postman, curl) during development
	if origin == "" {
		log.Println("No Origin header found, allowing connection (dev only).")
		return true
	}

	switch origin {
	case "ws://localhost:8082", "http://localhost:8080":
		return true
	default:
		log.Println("Blocked connection with Origin:", origin)
		return false
	}
}

func (m *Manager) Broadcast(msgType string, payload interface{}) {
	m.RLock()
	defer m.RUnlock()
	for client := range m.clients {
		client.egress <- Message{
			Type:    msgType,
			Payload: payload,
		}
	}
}

func (m *Manager) handleWebSocketRequest(c *Client, cudReq model.CUDRequest, successMsgType string, successMsg interface{}) error {
	responseChan := make(chan interface{})
	cudReq.ResponseChannel = responseChan
	m.UserService.QueueCUDRequest(cudReq)

	select {
	case response := <-responseChan:
		if err, ok := response.(error); ok {
			m.sendError(c.conn, successMsgType+"_response", err.Error())
			return err
		}
		m.sendSuccess(c.conn, successMsgType+"_response", successMsg)
	case <-time.After(5 * time.Second): // Timeout after 5 seconds
		m.sendError(c.conn, successMsgType+"_response", "Request timed out")
		return errors.New("request timed out")
	}
	return nil
}

func (m *Manager) handleCreateUser(message Message, c *Client) error {
	var req model.CreateUserRequest
	decodePayload(message.Payload, &req)
	cudReq := model.CUDRequest{
		Type:      "create_user",
		CreateReq: req,
	}
	return m.handleWebSocketRequest(c, cudReq, "create_user", "User created successfully")
}

func (m *Manager) handleGetUsers(_ Message, c *Client) error {
	cudReq := model.CUDRequest{
		Type: "get_users",
	}
	return m.handleWebSocketRequest(c, cudReq, "get_users", nil)
}

func (m *Manager) handleUpdateUser(message Message, c *Client) error {
	var req model.UpdateUserRequest
	decodePayload(message.Payload, &req)
	userID := message.Payload.(map[string]interface{})["user_id"].(int64)
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
	return m.handleWebSocketRequest(c, cudReq, "update_user", "User updated successfully")
}

func (m *Manager) handleDeleteUser(message Message, c *Client) error {
	userID := message.Payload.(map[string]interface{})["user_id"].(int64)
	cudReq := model.CUDRequest{
		Type:   "delete_user",
		UserID: userID,
	}
	return m.handleWebSocketRequest(c, cudReq, "delete_user", "User deleted successfully")
}

func (m *Manager) sendSuccess(conn *websocket.Conn, msgType string, data interface{}) {
	resp := Response{Type: msgType, Status: "success", Data: data}
	err := conn.WriteJSON(resp)
	if err != nil {
		return
	}
}

func (m *Manager) sendError(conn *websocket.Conn, msgType string, errMsg string) {
	resp := Response{Type: msgType, Status: "error", Error: errMsg}
	err := conn.WriteJSON(resp)
	if err != nil {
		return
	}
}

func decodePayload(input interface{}, out interface{}) {
	temp, _ := json.Marshal(input)   // change to map[string]interface{} -> json
	err := json.Unmarshal(temp, out) // change to json -> struct
	if err != nil {
		return
	}
}
