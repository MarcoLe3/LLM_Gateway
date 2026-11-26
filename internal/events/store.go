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
	_, err := s.DB.Exec(ctx,
		`INSERT INTO usage_events
		(id, caller_id, job_id, provider, model, input_tokens, output_tokens,
		cost_usd, latency_ms,cache_hit, success, error, created_at) 
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12, $13)
		ON CONFLICT (id) DO NOTHING`,
		e.ID, e.CallerID, e.JobID, e.Provider, e.Model, e.InputTokens, e.OutputTokens,
		e.CostUSD, e.LatencyMs, e.CacheHit, e.Success, e.Error, e.CreatedAt)

	return err
}
