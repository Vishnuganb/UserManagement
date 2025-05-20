package test_util

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketTestUtil struct {
	conn   *websocket.Conn
	queue  []map[string]interface{}
	mu     sync.Mutex
	closed bool
}

func NewWebSocketTestUtil(url string) (*WebSocketTestUtil, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	util := &WebSocketTestUtil{
		conn:   conn,
		queue:  make([]map[string]interface{}, 0),
		closed: false,
	}
	go util.listenForMessages()
	return util, nil
}

func (w *WebSocketTestUtil) listenForMessages() {
	for {
		if w.closed {
			return
		}

		_, message, err := w.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				panic("WebSocket read error: " + err.Error())
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal WebSocket message: %v", err)
			continue
		}
		w.mu.Lock()
		w.queue = append(w.queue, msg)
		w.mu.Unlock()
	}
}

// GetMessages retrieves and removes the next message from the queue
func (w *WebSocketTestUtil) GetMessages() (map[string]interface{}, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.queue) == 0 {
		return nil, false
	}

	msg := w.queue[0]
	w.queue = w.queue[1:]
	return msg, true
}

func (w *WebSocketTestUtil) Close() {
	w.closed = true
	if err := w.conn.Close(); err != nil {
		log.Printf("WebSocket close error: %v", err)
	}
}
