package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafka "github.com/segmentio/kafka-go"

	"UserManagement/internal/ws"
)

type KafkaEvent struct {
	Type    string      `json:"type"`    // e.g., user_created
	Payload interface{} `json:"payload"` // can be user ID, full user object, etc.
}

func StartConsumer(brokerAddr, topic string, manager *ws.Manager) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   topic,
		GroupID: "websocket-group",
	})

	go func() {
		defer func() {
			if err := reader.Close(); err != nil {
				log.Printf("Error closing Kafka reader: %v", err)
			}
		}()
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("Kafka consumer error:", err)
				continue
			}

			// Log the consumed message
			log.Printf("Consumed message: key=%s, value=%s", string(m.Key), string(m.Value))

			var event KafkaEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("Failed to unmarshal Kafka message: %v", err)
				continue
			}

			// Notify all clients when a user is created
			switch string(m.Key) {
			case "user_created":
				manager.Broadcast("New User Created", event.Payload)
			case "user_updated":
				manager.Broadcast("User Updated", event.Payload)
			case "user_deleted":
				manager.Broadcast("User Deleted", event.Payload)
			default:
				log.Printf("Unhandled Kafka event key: %s", m.Key)
			}
		}
	}()
}
