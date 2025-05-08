package kafka

import (
	"context"
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

func (p *Producer) Publish(key, value string) error {
	err := p.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	})
	if err != nil {
		log.Println("failed to publish message:", err)
	}
	return err
}

var _ service.MessageProducer = (*Producer)(nil)
