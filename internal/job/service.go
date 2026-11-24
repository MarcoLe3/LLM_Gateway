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

func NewStore(store *JobStore, pub queue.Queue, maxAttempts int, baseDelay time.Duration) *Service {
	return &Service{
		store:       store,
		pub:         pub,
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
	}
}

func (s *Service) SubmitJob(ctx context.Context, callerID uuid.UUID, key string, p RequestPayload) (*Job, error) {
	hash, err := p.Hash()
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:             uuid.New(),
		CallerID:       callerID,
		IdempotencyKey: key,
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
		existing, err := s.store.GetJobByIdempotencyKey(ctx, callerID, key)
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

func (s *Service) MarkJobFailed(ctx context.Context, job *Job, errMsg error) error {
	err := s.store.MarkRetry(ctx, job.ID, errMsg.Error())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) MarkJobDead(ctx context.Context, job *Job, errMsg error) error {
	err := s.store.MarkDead(ctx, job.ID, errMsg.Error())
	if err != nil {
		return err
	}
	
	return nil
}
