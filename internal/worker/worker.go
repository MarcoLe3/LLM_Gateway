package worker

import (
	"context"
	"llm_gateway/internal/config"
	"llm_gateway/internal/job"
	"llm_gateway/internal/queue"
	"llm_gateway/internal/router"

	"github.com/google/uuid"
)

type Worker struct {
	svc    *job.Service
	router *router.Router
	queue  queue.Queue
	cfg    config.Config
}

func (w *Worker) processJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := w.svc.ClaimJobToRun(ctx, jobID)
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	res, err := w.router.Complete(ctx, job.Payload)
	if err != nil {
		return w.svc.MarkJobFailed(ctx, job, err)
	}

	return w.svc.MarkSucceeded(ctx, jobID, res)
}

func (w *Worker) Run(ctx context.Context) error {
	return w.queue.Consume(ctx, w.cfg.Workers, w.processJob)
}
