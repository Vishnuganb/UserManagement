package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

// ClientList is a map of clients to their connection status
type ClientList map[*Client]bool

// Client represents a single client connection
type Client struct {
	conn    *websocket.Conn
	manager *Manager
	egress  chan Message //[]byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan Message),
	}
}

func (c *Client) readMessages() {
	defer func() {
		// clean up connection
		c.manager.removeClient(c)
	}()
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("Error setting read deadline:", err)
		return
	}

	// sometimes they will send really big message
	c.conn.SetReadLimit(512)

	// whenever we receive a pong message it will trigger the func that we assign
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var request Message
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshaling event : %v", err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Printf("error handling message : %v", err)
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop() // Stop the ticker when the function exits

	for {
		select {
		case message, ok := <-c.egress: // reading
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Printf("connection closed: %v", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("error marshaling event : %v", err)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to send message: %v", err)
				return
			}
			log.Printf("message sent")

		case <-ticker.C: // receive a value from the channel
			log.Println("ping")

			// send a Ping to the client
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(_ string) error {
	log.Println("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
