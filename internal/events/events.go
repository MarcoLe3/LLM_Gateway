package events

import (
	"context"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, key string, payload ) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key: 
	})
}

func (p *KafkaProducer) Close() {
	return p.writer.Close()
}
