package provider

import (
	"context"
	"llm_gateway/internal/jobs"
	"net/http"
)

type Claude struct {
	apiKey string
	model  string
	client *http.Client
}

type claudeRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []jobs.Message
}

func NewClaude(apiKey string, model string) *Claude {
	return &Claude{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

func (c *Claude) Name() string {
	return "claude"
}

func (c *Claude) Complete(ctx context.Context, p jobs.RequestPayload) (jobs.ResponsePayload, error) {

}
