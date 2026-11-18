package job

import (
	"uuid"
	"time"
)

type State string

const (
	StateQueued 	State = "queued"
	StateRunning	State = "running"
	StateSuccessed	State = "succeeded"
	StateFailed		State = "failed"
	StateDead		State = "dead"
)

type Job struct {
	ID			uuid.UUID
	CallerID	uuid.UUID
	LatencyMs	int
	UpatedAt	*time.Time
	CreatedAt	*time.Time
	StartedAt	*time.Time
	FinishedAt	*time.Time
}