package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Queue interface {
	Publish(ctx context.Context, jobID uuid.UUID) error
	PublishDelayed(ctx context.Context, jobID uuid.UUID, d time.Duration) error
	PublishDead(ctx context.Context, jobID uuid.UUID) error
	Consume(ctx context.Context, jobID uuid.UUID) error
	Close() error
}
