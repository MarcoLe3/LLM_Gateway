package worker

import (
	"context"
	"llm_gateway/internal/config"
	"llm_gateway/internal/job"
	"llm_gateway/internal/queue"
	"llm_gateway/internal/router"
	"uuid"
)

type Worker struct {
	svc    *job.Service
	router *router.Router
	queue  queue.Queue
	cfg    config.Config
}

func (w *Worker) processJob(ctx context.Context, jobID uuid.UUID) {
	job, err := w.svc.
}
