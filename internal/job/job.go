package job

import (
	"time"

	"github.com/google/uuid"
)

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSuccessed State = "succeeded"
	StateFailed    State = "failed"
	StateDead      State = "dead"
)

type Job struct {
	ID             uuid.UUID       `json:"id"`
	CallerID       uuid.UUID       `json:"-"`
	IdempotencyKey string          `json:"-"`
	State          State           `json:"state"`
	Attempts       int             `json:"attempts"`
	Provider       string          `json:"provider,omitempty"`
	Model          string          `json:"model,omitempty"`
	InputTokens    int             `json:"input_tokens,omitempty"`
	OutputTokens   int             `json:"output_tokens,omitempty"`
	RequestHash    string          `json:"request_hash"`
	Payload        RequestPayload  `json:"-"`
	Result         ResponsePayload `json:"result,omitempty"`
	Error          string          `json:"-"`
	NextRunAt      *time.Time      `json:"-"`
	UpdatedAt      *time.Time      `json:"-"`
	CreatedAt      *time.Time      `json:"created_at"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
}
