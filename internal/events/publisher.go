package events

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"

	"llm_gateway/internal/job"
)

type Publisher interface {
	Publish(ctx context.Context, e UsageEvent) error 
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
}

const Topic = "usage events"

func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    Topic,
			Balancer: &kafka.LeastBytes{},
			Async: true
		},
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, key string, e UsageEvent) error {
	b, err := json.Marshal(e)

	if err != nil {
		return err
	}
	
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(key),
		Value: b,
	})
}

func (p *KafkaProducer) Close() {
	return p.writer.Close()
}
