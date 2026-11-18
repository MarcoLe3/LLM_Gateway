package events

import (
	"time"

	"github.com/google/uuid"
)

type UsageEvent struct {
	ID           uuid.UUID  `json:"id"`
	CallerID     uuid.UUID  `json:"caller_id"`
	JobID        *uuid.UUID `json:"job_id, omitempty"`
	Provder      string     `json:"provider"`
	Model        string     `json:"model"`
	InputTokens  int        `json:"input_tokens"`
	OutputTokens int        `json:"output_tokens"`
	CostUSD      float64    `json:"cost_usd"`
	LatencyMs    int        `json:"latency_ms"`
	CacheHit     bool       `json:"cache_hit"`
	Success      bool       `json:"success"`
	Error        string     `json:"error"`
	CreatedAt    time.Time  `json:"created_at"`
}
