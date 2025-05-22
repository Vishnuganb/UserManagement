package kafka

import (
	"context"
	"encoding/json"
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

			// Notify all clients when a user is created
			manager.Broadcast(string(m.Key), json.RawMessage(m.Value))
		}
	}()
}
