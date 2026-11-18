package events

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	store  EventStore
}

func NewKafkaConsumer(brokers []string, groupID string, store EventStore) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		store: store,
	}
}

func (c *KafkaConsumer) Run(ctx context.Context) error {
	defer c.reader.Close()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		var e UsageEvent
		if err = json.Unmarshal(msg.Value, &e); err != nil {
			return err
		}
	}
}

func (c *KafkaConsumer) insertWithRetry(ctx context.Context, e UsageEvent) {
	var err error
}
