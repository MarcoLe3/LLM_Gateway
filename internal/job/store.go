package job

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStore struct {
	DB *pgxpool.Pool
}

func NewJobStore(pool *pgxpool.Pool) *JobStore {
	return &JobStore{
		DB: pool,
	}
}

func (s *JobStore) InsertJob(ctx context.Context, j *Job) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO jobs
	(id, caller_id,latency_ms, updated_at, created_at, started_at, finished_at)
	ON VALUES ($1,$2,$3,$4, $5,$6,$7)`, 
	j.ID, j.CallerID, j.LatencyMs, j.UpatedAt, j.CreatedAt, j.StartedAt, j.FinishedAt)
	return err
}
