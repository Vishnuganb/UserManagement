package ws

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

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
	log.Println("new connection")
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

func (m *Manager) Broadcast(msgType string) {
	m.RLock()
	defer m.RUnlock()
	for client := range m.clients {
		client.egress <- Message{
			Type:    msgType,
			Payload: "success",
		}
	}
}

func (m *Manager) handleCreateUser(message Message, c *Client) error {
	var req model.CreateUserRequest
	decodePayload(message.Payload, &req)
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "create_user",
		CreateReq:       req,
		ResponseChannel: responseChan,
	}
	m.UserService.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		m.sendError(c.conn, "get_user_response", err.Error())
		return err
	}
	//m.Broadcast("user_list_updated", "A new user was created")
	return nil
}

func (m *Manager) handleGetUsers(_ Message, c *Client) error {
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "get_users",
		ResponseChannel: responseChan,
	}
	m.UserService.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		m.sendError(c.conn, "get_user_response", err.Error())
		return err
	}
	m.sendSuccess(c.conn, "get_users_response", response)
	return nil
}

func (m *Manager) handleUpdateUser(message Message, c *Client) error {
	var req model.UpdateUserRequest
	decodePayload(message.Payload, &req)
	userID := message.Payload.(map[string]interface{})["user_id"].(int64)
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
	m.UserService.QueueCUDRequest(cudReq)
	response := <-responseChan
	if err, ok := response.(error); ok {
		m.sendError(c.conn, "update_user_response", err.Error())
		return err
	}
	m.sendSuccess(c.conn, "update_user_response", "User updated successfully")
	return nil
}

func (m *Manager) handleDeleteUser(message Message, c *Client) error {
	userID := message.Payload.(map[string]interface{})["user_id"].(int64)
	responseChan := make(chan interface{})
	cudReq := model.CUDRequest{
		Type:            "delete_user",
		UserID:          userID,
		ResponseChannel: responseChan,
	}
	m.UserService.QueueCUDRequest(cudReq)
	response := <-responseChan
	// Check if the response is an error
	if err, ok := response.(error); ok {
		m.sendError(c.conn, "delete_user_response", err.Error())
		return err
	}
	// If not an error, send success
	m.sendSuccess(c.conn, "delete_user_response", "User deleted successfully")
	return nil
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
