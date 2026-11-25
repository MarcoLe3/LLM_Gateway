package job

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStore struct {
	DB *pgxpool.Pool
}

const jobCols = `id, caller_id, idempotency_key, request_hash, state, attempts, 
COALESCE(provider, ''), COALESCE(model, ''), request_hash, request_payload, response_payload,
COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), COALESCE(last_error, ''),
created_at, started_at, updated_at, finished_at`

func NewJobStore(pool *pgxpool.Pool) *JobStore {
	return &JobStore{
		DB: pool,
	}
}

func scanJob(row pgx.Row) (*Job, error) {
	var j Job
	err := row.Scan(
		&j.ID, &j.CallerID, &j.IdempotencyKey, &j.State, &j.Attempts,
		&j.Provider, &j.Model, &j.Payload, &j.Result,
		&j.InputTokens, &j.OutputTokens, &j.Error,
		&j.CreatedAt, &j.StartedAt, &j.UpdatedAt, &j.FinishedAt,
	)

	if err != nil {
		return nil, err
	}

	return &j, nil
}

func requireOne(tag pgconn.CommandTag, err error) error {
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return ErrStateConflicted
	}

	return nil
}

func (s *JobStore) InsertJob(ctx context.Context, j *Job) (bool, error) {
	tag, err := s.DB.Exec(ctx, `
		INSERT INTO jobs
			(id, caller_id, idempotency_key, updated_at, request_hash, state, request_payload)
		ON VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (caller_id, idempotency_key) DO NOTHING`,
		j.ID, j.CallerID, j.IdempotencyKey, j.RequestHash, j.State, j.Payload)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

func (s *JobStore) GetJobByID(ctx context.Context, callerID, id uuid.UUID) (*Job, error) {
	return scanJob(s.DB.QueryRow(ctx, `SELECT`+jobCols+`FROM jobs WHERE caller_id = $1
	AND id = $2`, callerID, id))
}

func (s *JobStore) GetJobByIdempotencyKey(ctx context.Context, callerID uuid.UUID, key string) (*Job, error) {
	return scanJob(s.DB.QueryRow(ctx, `SELECT`+jobCols+`FROM jobs WHERE caller_id = $1
	AND idempotency_key = $2`, callerID, key))
}

func (s *JobStore) ClaimJobToRun(ctx context.Context, id uuid.UUID) (*Job, error) {
	return scanJob(s.DB.QueryRow(ctx, `
		UPDATE jobs
		SET state = 'running', attempts = attempts + 1,
			started_at = now(), updated_at = now()
		WHERE id = $1 AND state = 'queued'
		RETURNING`+jobCols, id))
}

func (s *JobStore) MarkSucceeded(ctx context.Context, id uuid.UUID, r *ResponsePayload) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE jobs
		SET state = 'succeeded', response_payload = $2, provider = $3,
			input_tokens = $4, output_tokens = $5,
			finished_at = now(), updated_at = now()
		WHERE id = $1 and state = 'running'`,
		id, r, r.Provider, r.InputTokens, r.OutputTokens)
	return requireOne(tag, err)
}

func (s *JobStore) MarkDead(ctx context.Context, id uuid.UUID, errMsg string) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE jobs
		SET state = 'dead', last_error = $2, last_error_at = now(),
			finished_at = now(), updated_at = now()
		WHERE id = $1 and state = 'running'`,
		id, errMsg)
	return requireOne(tag, err)
}

func (s *JobStore) MarkRetry(ctx context.Context, id uuid.UUID, errMsg string, delay time.Time) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE jobs
		SET state = 'queued', last_error = $2, last_error_at = now(),
			updated_at = now()
			next_run_at = now() + make_interval(secs => $3)
		WHERE id = $1 and state = 'running'`,
		id, errMsg, delay)
	return requireOne(tag, err)
}
