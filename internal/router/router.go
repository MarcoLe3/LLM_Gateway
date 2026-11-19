package router

import (
	"context"
	"llm_gateway/internal/jobs"
	"llm_gateway/internal/provider"
	"math/rand"
	"time"
)

type Router struct {
	providers   []provider.Provider
	maxAttempts int
	timeout     time.Duration
}

func (r *Router) Complete(ctx context.Context, p jobs.RequestPayload) {

}

func (r *Router) backoff(ctx context.Context, attempt int) error {
	d := time.Duration(1 << attempt)
	d = d/2 + time.Duration(rand.Int63n(int64(d/2)+1))
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
