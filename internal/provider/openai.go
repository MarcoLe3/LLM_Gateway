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

type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

type openRequest struct {
	Model     string        `json:"model"`
	Messages  []job.Message `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
	Tempature *float64      `json:"tempature,omitempty"`
}

type openResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

func (o *OpenAI) Complete(ctx context.Context, p job.RequestPayload) (*job.ResponsePayload, error) {
	body, err := json.Marshal(openRequest{
		Model:     p.Model,
		Messages:  p.Messages,
		MaxTokens: p.MaxTokens,
		Tempature: p.Tempature,
	})

	if err != nil {
		return nil, err
	}

	const url = "https://api.openai.com/v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer"+o.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, &Error{Provider: "openai", Status: resp.StatusCode, Err: fmt.Errorf("%s", raw)}
	}

	var openResp openResponse
	if err = json.Unmarshal(raw, &openResp); err != nil || len(openResp.Choices) == 0 {
		return nil, &Error{Provider: "openai", Status: 502, Err: fmt.Errorf("%s", raw)}
	}

	return &job.ResponsePayload{
		Text:         openResp.Choices[0].Message.Content,
		InputTokens:  openResp.Usage.PromptTokens,
		OutputTokens: openResp.Usage.CompletionTokens,
		Provider:     "openai",
	}, nil
}
