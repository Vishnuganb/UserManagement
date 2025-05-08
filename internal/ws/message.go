package ws

// Message Client request message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MessageHandler func(message Message, c *Client) error

// Response message
type Response struct {
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}
