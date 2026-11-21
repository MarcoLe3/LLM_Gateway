package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"llm_gateway/internal/job"
	"net/http"
)

type Claude struct {
	apiKey string
	model  string
	client *http.Client
}

type claudeRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
	Messages  []job.Message `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`

	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func NewClaude(apiKey, model string) *Claude {
	return &Claude{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

func (c *Claude) Name() string {
	return "claude"
}

func (c *Claude) Complete(ctx context.Context, p job.RequestPayload) (*job.ResponsePayload, error) {
	var systemMsg string
	msgs := make([]job.Message, 0, len(p.Messages))
	for _, m := range p.Messages {
		if m.Role == "system" {
			systemMsg = m.Content
			continue
		}
		msgs = append(msgs, m)
	}

	maxTokens := p.MaxTokens
	if maxTokens <= 0 || maxTokens > 1024 {
		maxTokens = 1024
	}

	body, _ := json.Marshal(claudeRequest{
		Model:     p.Model,
		MaxTokens: maxTokens,
		System:    systemMsg,
		Messages:  msgs,
	})

	const url = "https://api.anthropic.com/v1/messages"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("context-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, &Error{Provider: "claude", Status: resp.StatusCode, Err: fmt.Errorf("%s", err)}
	}

	var output claudeResponse
	if err = json.Unmarshal(raw, &output); err != nil || len(output.Content) == 0 {
		return nil, &Error{Provider: "claude", Status: 502, Err: fmt.Errorf("bad response")}
	}

	return &job.ResponsePayload{
		Text:         output.Content[0].Text,
		InputTokens:  output.Usage.InputTokens,
		OutputTokens: output.Usage.OutputTokens,
		Provider:     "claude",
	}, nil
}
