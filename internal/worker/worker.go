package worker

import (
	"context"
	"llm_gateway/internal/config"
	"llm_gateway/internal/events"
	"llm_gateway/internal/job"
	"llm_gateway/internal/queue"
	"llm_gateway/internal/router"
	"time"

	"github.com/google/uuid"
)

type jobService interface {
	ClaimJobToRun(ctx context.Context, jobID uuid.UUID) (*job.Job, error)
	MarkSucceeded(ctx context.Context, jobID uuid.UUID, res *job.ResponsePayload) error
	MarkJobFailed(ctx context.Context, j *job.Job, errMsg error) error
}

type rtr interface {
	Complete(ctx context.Context, p job.RequestPayload) (*job.ResponsePayload, error)
}

type Worker struct {
	svc    jobService
	router rtr
	queue  queue.Queue
	events events.Publisher
	cfg    config.Config
}

func NewWorker(svc *job.Service,
	r *router.Router, q queue.Queue, events events.Publisher, cfg config.Config) *Worker {
	return &Worker{
		svc:    svc,
		router: r,
		queue:  q,
		events: events,
		cfg:    cfg,
	}
}

func (w *Worker) processJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := w.svc.ClaimJobToRun(ctx, jobID)
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	start := time.Now()
	latency := int(time.Since(start).Milliseconds())

	res, err := w.router.Complete(ctx, job.Payload)
	if err != nil {
		w.events.Publish(ctx, events.UsageEvent{
			CallerID: job.CallerID, JobID: &job.ID, Success: false,
			Error: err.Error(), LatencyMs: latency,
		})
		return w.svc.MarkJobFailed(ctx, job, err)
	}

	if err := w.svc.MarkSucceeded(ctx, jobID, res); err != nil {
		return err
	}
	w.events.Publish(ctx, events.UsageEvent{
		CallerID: job.CallerID, JobID: &job.ID, Provider: res.Provider,
		InputTokens: res.InputTokens, OutputTokens: res.OutputTokens,
		Success: true, LatencyMs: latency,
	})

	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	return w.queue.Consume(ctx, w.cfg.Workers, w.processJob)
}
