package queue

import (
	"context"

	"github.com/google/uuid"
)

type Queue interface {
	Publish(ctx context.Context, jobID uuid.UUID) error
	Consume(ctx context.Context, jobID uuid.UUID) error
}