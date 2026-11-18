package provider

import (
	"context"
	
	"llm_gateway/internal/jobs"
)

type Provider interface {
	Complete(ctx context.Context, payload jobs.RequestPayload) (*Request, error)
}
