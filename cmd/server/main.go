package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"UserManagement/internal/handler"
	"UserManagement/internal/kafka"
	"UserManagement/internal/router"
	"UserManagement/internal/service"
	"UserManagement/internal/util"
	"UserManagement/internal/validator"
	"UserManagement/internal/ws"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to the database:", err)
	}

	defer func(conn *sql.DB) {
		_ = conn.Close()
	}(conn)

	ctx := context.Background()
	v := validator.NewValidator()
	producer := kafka.NewProducer(config.KafkaBroker, config.KafkaTopic)
	us := service.NewUserService(ctx, conn, v, producer)
	uh := handler.NewUserHandler(us)
	r := router.NewRouter(uh)

	// WebSocket setup
	m := ws.NewManager(us)

	// Start Kafka consumer
	go kafka.StartConsumer(config.KafkaBroker, config.KafkaTopic, m)

	// Run REST API server on port 8080
	go func() {
		log.Println("Starting REST server at :8080")
		err = http.ListenAndServe(":8080", r)
		if err != nil {
			log.Fatal("error starting REST server:", err)
		}
	}()


	http.HandleFunc("/ws_users", m.ServeWS) // Handle WebSocket connection
	log.Println("Starting WebSocket server at :8082") // Run WebSocket server on port 8082
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal("error starting WebSocket server:", err)
	}
}
