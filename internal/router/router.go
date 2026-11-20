package router

import (
	"context"
	"errors"
	"fmt"
	"llm_gateway/internal/job"
	"llm_gateway/internal/provider"
	"log/slog"
	"math/rand"
	"time"
)

type Router struct {
	providers   []provider.Provider
	maxAttempts int
	timeout     time.Duration
}

func (r *Router) Complete(ctx context.Context, p job.RequestPayload) (*job.ResponsePayload, error) {
	var errs []error
	for _, prov := range r.providers {
		for attempt := 0; attempt < r.maxAttempts; attempt++ {
			callCtx, cancel := context.WithTimeout(ctx, r.timeout)
			res, err := prov.Complete(callCtx, p)
			cancel()

			if err == nil {
				return res, nil
			}

			slog.Warn("provider attempt failed", "provider",
				prov.Name(), "attempt", attempt+1, "err", err)

			errs = append(errs, fmt.Errorf("%s attempt %d: %w", prov.Name(), attempt+1, err))

			if !r.canRetry(err) {
				return nil, errors.Join(errs...)
			}

			if attempt < r.maxAttempts-1 {
				if err := r.backoff(ctx, attempt); err != nil {
					return nil, errors.Join(errs...)
				}
			}
		}
	}

	return nil, fmt.Errorf("all providers failed:\n %w", errors.Join(errs...))
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

func (r *Router) canRetry(err error) bool {
	var pe *provider.Error
	if errors.As(err, pe) {
		return pe.Status == 420 || pe.Status >= 500
	}

	return true
}
