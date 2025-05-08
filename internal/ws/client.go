package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second    // if i send the ping i will wait max 10 s before i drop the connection
	pingInterval = (pongWait * 9) / 10 // how often we will send ping to the client, it needs to lower than pong wait
)

type ClientList map[*Client]bool

/*
Even though Manager creates the WebSocket conn, it does not store it.
Instead:

	The Manager creates the conn.

	Then passes it to the Client constructor.

	The Client stores and owns the conn.

Why not store conn in Manager?

	Because Manager isn’t meant to manage individual connection logic. It only:

		* Accepts new connections.

		* Maybe keeps track of all clients.

		* Routes messages or commands.

Why Client must store conn:

	The Client is responsible for:

		* Reading from conn (conn.ReadMessage()).

		* Writing to conn (conn.WriteMessage()).

		* Handling disconnection.

		* Responding to pings/pongs, etc.

Why egress chan []byte is added in Client, not Manager?

Short answer:

	Because each client has its own WebSocket connection and its own messages to send, so it needs its own outgoing channel (egress).

The Manager manages many clients — it doesn't deal directly with reading/writing messages.

The Client owns the WebSocket connection (conn) and is responsible for:

  - Reading messages (from the WebSocket)

  - Writing messages (to the WebSocket)

Writing to a WebSocket must be done carefully — only one goroutine should write at a time.

That’s why each Client has an egress channel:

It’s a safe and controlled way to queue messages to be sent to that specific client.

If Manager had one big channel for all clients, things would get mixed up and unsafe.
*/
/*
	Client creates a JSON object:


	{
	  "type": "send_message",
	  "payload": {
		"text": "Hi"
	  }
	}

	Then it stringifies it:

	JSON.stringify({...})

	Result:

	'{"type":"send_message","payload":{"text":"Hi"}}'

	That string is sent over the WebSocket.

	On the client side (e.g., JavaScript):

	socket.send('{"type":"send_message","payload":{"text":"Hi"}}');

	This looks like a string to you.

	But under the hood:

	The browser converts that string into bytes (using UTF-8 encoding).

	Then sends those bytes over the WebSocket connection.

	On the Go server side, you receive it as []byte.

	You can then:

	Log it: log.Println(string(payload))

	Unmarshal it into a struct like:

	var event Event
	json.Unmarshal(payload, &event)
*/
type Client struct {
	conn    *websocket.Conn
	manager *Manager

	// egress is used to avoid concurrent writes on the websocket connection
	/*
		Why []byte (byte slice) is used?
			WebSocket messages are received and sent as binary data, and in Go, that’s a []byte.

			So:

				When a client sends a message → it comes in as []byte

				When you want to send it back → you also use []byte

				Later, you convert []byte to string or decode it as JSON if needed:

				log.Printf("Payload: %v", string(payload))
	*/
	/*
		When we say WebSocket messages are sent/received as []byte, it can contain anything:

		It could be plain text (like a string or JSON)

		Or actual binary (like images, files, etc.)

		But it's always handled as []byte in Go — that’s why we say it's "binary data".
	*/
	/*
		Why move to using Event?

			Instead of just relaying bytes, the server can now:

			Understand: What kind of action is the client asking for? (send_message, join_room, etc.)

			Route: Run the correct function based on that type.

			Validate: Make sure the data inside is correct (e.g., the text isn't empty).

			Extend: Easily support new actions later without rewriting everything.
	*/
	egress chan Message //[]byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan Message), // egress is a channel used to send messages out of the client (from server → client).
	}
}

func (c *Client) readMessages() {
	// defer tells Go to run a function only when the surrounding function ends —
	// even if it ends because of an error, return, or panic.
	/*
		1. defer:
			Means "run this later, when readMessages() is done."

		2. func() { ... }:
			Defines an anonymous function (a function with no name).

		3. The () at the end:
			That calls the anonymous function immediately, but defer delays its actual execution.

		So you're defining and scheduling this anonymous function to run later.
	*/
	defer func() {
		// clean up connection
		c.manager.removeClient(c)
	}()
	// setReadLine using for how long we need to wait, that's why we are adding 10s with current time
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("Error setting read deadline:", err)
		return
	}

	// sometimes they will send really big message
	c.conn.SetReadLimit(512)

	// whenever we receive a pong message it will trigger the func that we assign
	c.conn.SetPongHandler(c.pongHandler)

	// This is a "forever loop" (infinite loop). It's common in long-lived connections like WebSockets,
	// where the server is meant to keep listening for messages until something breaks or disconnects.
	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			// connections is closed without client or server is sending connection close message it's called the abnormal way
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		//// whenever it's reading message it is also broadcast to all other clients
		//for wsClient := range c.manager.clients {
		//	wsClient.egress <- payload // Write to the channel
		//}
		//
		///*
		//	Code							Meaning
		//	<-channel					Read from the channel
		//	channel <- value			Write to the channel
		//
		//	1 = Text Message
		//
		//	2 = Binary Message
		//
		//	8 = Close
		//
		//	9 = Ping
		//
		//	10 = Pong
		//
		//	These numbers are standardized in the RFC.
		//*/
		//log.Printf("readMessageType: %v", messageType)
		//log.Printf("readMessagePayload: %v", string(payload)) // payload is coming as a binary so convert to string
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
			}
			log.Printf("message sent")

		/*
			What happens:
				The loop runs.

				The first 9 seconds, nothing happens unless there's a message.

				When 10 seconds pass → ticker.C sends a signal.

				Then case <-ticker.C: runs:

				log.Println("ping")
				conn.WriteMessage(websocket.PingMessage, [])
				Another 10 seconds go by...

				It pings again.
		*/
		case <-ticker.C: // receive a value from the channel
			log.Println("ping")

			// send a Ping to the client
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
