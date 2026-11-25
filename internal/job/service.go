package job

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"llm_gateway/internal/queue"
)

type Service struct {
	store       *JobStore
	pub         queue.Queue
	maxAttempts int
	baseDelay   time.Duration
}

func NewService(store *JobStore, pub queue.Queue, maxAttempts int, baseDelay time.Duration) *Service {
	return &Service{
		store:       store,
		pub:         pub,
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
	}
}

func (s *Service) SubmitJob(ctx context.Context, callerID uuid.UUID, Idemkey string, p RequestPayload) (*Job, error) {
	hash, err := p.Hash()
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:             uuid.New(),
		CallerID:       callerID,
		IdempotencyKey: Idemkey,
		RequestHash:    hash,
		State:          StateQueued,
		Payload:        p,
	}

	inserted, err := s.store.InsertJob(ctx, job)
	if err != nil {
		return nil, err
	}

	if !inserted {
		// check if job already exist instead if we couldnt insert it
		existing, err := s.store.GetJobByIdempotencyKey(ctx, callerID, Idemkey)
		if err != nil {
			return nil, err
		}

		if existing.RequestHash != hash {
			return nil, ErrKeyReuse
		}

		return existing, err
	}

	if err = s.pub.Publish(ctx, job.ID); err != nil {
		slog.Error("publish failed")
	}

	return job, nil
}

func (s *Service) GetJob(ctx context.Context, callerID, jobID uuid.UUID) (*Job, error) {
	job, err := s.store.GetJobByID(ctx, callerID, jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Service) ClaimJobToRun(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	job, err := s.store.ClaimJobToRun(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job, err
}

func (s *Service) MarkSucceeded(ctx context.Context, jobID uuid.UUID, res *ResponsePayload) error {
	return s.store.MarkSucceeded(ctx, jobID, res)
}

func (s *Service) MarkJobFailed(ctx context.Context, job *Job, errMsg error) error {
	msg := truncate(errMsg.Error(), 2000)

	if job.Attempts >= s.maxAttempts {
		if err := s.store.MarkDead(ctx, job.ID, msg); err != nil {
			return err
		}

		return s.pub.PublishDead(ctx, job.ID)
	}

	delay := queue.RetryDelay(job.Attempts)
	if err := s.store.MarkRetry(ctx, job.ID, msg, delay); err != nil {
		return err
	}
	return s.pub.PublishDelayed(ctx, job.ID, job.Attempts)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}

	return s
}
