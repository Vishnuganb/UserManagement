package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafka "github.com/segmentio/kafka-go"

	"UserManagement/internal/service"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokerAddr string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) NotifyUserCreated(key string, value interface{}) error {
	// Serialize the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		log.Println("failed to serialize value:", err)
		return err
	}
	err = p.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
	if err != nil {
		log.Println("failed to publish message:", err)
	}
	return err
}

var _ service.UserNotifier = (*Producer)(nil)
