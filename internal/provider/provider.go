package provider

import (
	"context"
	"fmt"

	"llm_gateway/internal/jobs"
)

type Provider interface {
	Name() string
	Complete(ctx context.Context, payload jobs.RequestPayload) (*jobs.ResponsePayload, error)
}

type Error struct {
	Provider string
	Status   int
	Err      error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: status %d: %v", e.Provider, e.Status, e.Err)
}


