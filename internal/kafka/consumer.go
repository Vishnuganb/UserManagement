package kafka

import (
	"context"
	"log"

	kafka "github.com/segmentio/kafka-go"

	"UserManagement/internal/ws"
)

func StartConsumer(brokerAddr, topic string, manager *ws.Manager) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   topic,
		GroupID: "websocket-group",
	})

	go func() {
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("Kafka consumer error:", err)
				continue
			}

			// Log the consumed message
			log.Printf("Consumed message: key=%s, value=%s", string(m.Key), string(m.Value))

			// Notify all clients when a user is created
			if string(m.Key) == "user_created" {
				manager.Broadcast("New User Created")
			} else if string(m.Key) == "user_deleted" {
				manager.Broadcast("User Updated")
			} else if string(m.Key) == "user_updated" {
				manager.Broadcast("User Deleted")
			}
		}
	}()
}
