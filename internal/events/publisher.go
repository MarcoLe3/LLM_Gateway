package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Publisher interface {
	Publish(ctx context.Context, e UsageEvent) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			Async:        true,
			BatchTimeout: 50 * time.Millisecond,
			Completion: func(msgs []kafka.Message, err error) {
				if err != nil {
					slog.Error("kafka failed to publish", "err", err, "count", len(msgs))
				}
			},
		},
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, e UsageEvent) error {
	b, err := json.Marshal(e)

	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(e.CallerID.String()),
		Value: b,
	})
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
