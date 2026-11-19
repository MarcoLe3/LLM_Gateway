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
	ID         uuid.UUID `json:"id"`
	CallerID   uuid.UUID `json:"caller_id"`
	LatencyMs  int       `json:"latency_ms"`
	UpatedAt   *time.Time
	CreatedAt  *time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}
