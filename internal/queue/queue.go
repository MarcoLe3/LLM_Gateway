package queue

import (
	"context"

	"github.com/google/uuid"
)

type Queue interface {
	Publish(ctx context.Context, jobID uuid.UUID) error
	PublishDelayed(ctx context.Context, jobID uuid.UUID, attempt int) error
	PublishDead(ctx context.Context, jobID uuid.UUID) error
	Consume(ctx context.Context, concurrency int, h Handler) error
	Close() error
}

type Handler func(ctx context.Context, jobID uuid.UUID) error
