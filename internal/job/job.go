package job

import (
	"time"
	"uuid"
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
	ID             uuid.UUID  `json:"id"`
	CallerID       uuid.UUID  `json:"-"`
	IdempotencyKey uuid.UUID  `json:"-"`
	State          State      `json:"state"`
	Attempts       int        `json:"attempts"`
	Provider       string     `json:"provider,omitempty"`
	Error          string     `json:"-"`
	UpdatedAt      *time.Time `json:"-"`
	CreatedAt      *time.Time `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}
