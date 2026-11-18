package events

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventStore struct {
	DB *pgxpool.Pool
}

func NewEventStore(pool *pgxpool.Pool) *EventStore {
	return &EventStore{
		DB: pool,
	}
}

func (s *EventStore) Insert(ctx context.Context, e UsageEvent) error {
	
}