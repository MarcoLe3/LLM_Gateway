package limiter

import (
	"context"
	"time"
	"uuid"

	"github.com/redis/go-redis/v9"
)

var luaSrc string

type Limiter struct {
	rdb    *redis.Client
	script *redis.Script
}

type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{
		rdb:    rdb,
		script: redis.NewScript(luaSrc),
	}
}

func (l *Limiter) Allow(ctx context.Context, callerID uuid.UUID, rpm int) (Result, error) {
	rate := float64(rpm)
	key := "rl:" + callerID.String()

	vals, err := l.script.Run(ctx, l.rdb, []string{key}, rpm, rate, 120).Slice()
}
