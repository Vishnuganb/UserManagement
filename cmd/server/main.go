package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"UserManagement/internal/handler"
	"UserManagement/internal/kafka"
	"UserManagement/internal/router"
	"UserManagement/internal/service"
	"UserManagement/internal/validator"
	"UserManagement/internal/ws"
)

const (
	dbDriver   = "postgres"
	dbSource   = "postgresql://root:secret@localhost:5433/userManagement?sslmode=disable"
	brokerAddr = "localhost:9092"
	topic      = "user_topic"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to the database:", err)
	}
	// Run this line at the end of the function
	defer func(conn *sql.DB) {
		_ = conn.Close()
	}(conn) // When the main() function finishes then automatically close the connection.

	v := validator.NewValidator()
	producer := kafka.NewProducer(brokerAddr, topic)
	us := service.NewUserService(conn, v, producer)
	uh := handler.NewUserHandler(us)
	r := router.NewRouter(uh)

	// WebSocket setup
	m := ws.NewManager(us)

	// Start Kafka consumer
	go kafka.StartConsumer(brokerAddr, topic, m)

	// Run REST API server on port 8080
	/*
		By using go, we’re saying:

		“Start the REST server in the background — don’t wait for it — and move on to start the WebSocket server too.”
	*/
	go func() {
		log.Println("Starting REST server at :8080")
		err = http.ListenAndServe(":8080", r)
		if err != nil {
			log.Fatal("error starting REST server:", err)
		}
	}()

	// Run WebSocket server on port 8082
	/*
		1. In your Go server, you only say:

		http.HandleFunc("/ws", wh.HandleWebSocket)
		http.ListenAndServe(":8082", nil)
		This means:

		“I’m listening for requests on /ws over HTTP (on port 8082).”

		2. In your frontend code (or Postman WebSocket tab), you write:

		const socket = new WebSocket("ws://localhost:8082/ws");
		The ws:// tells the browser:

		“Hey, this is a WebSocket connection.”

		Behind the scenes, the browser makes an HTTP request and adds a special header like this:


		Upgrade: websocket
		This tells the server:

		“Please switch this from HTTP to WebSocket.”

		And your Go server upgrades the connection using:

		conn, err := upgrader.Upgrade(w, r, nil)
	*/
	/*
		In REST (your project), you're not using http.HandleFunc because you're
		using a router library (like chi or gorilla/mux) to handle your routes
		more cleanly.

	*/
	// This registers your WebSocket handler in the default mux.
	http.HandleFunc("/ws_users", m.ServeWS) // Handle WebSocket connection
	log.Println("Starting WebSocket server at :8082")
	// Start an HTTP server on port 8082, and use the default mux (which has the /ws route I just added).
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal("error starting WebSocket server:", err)
	}
}
