package auth

import (
	"context"
	"uuid"
)

type Caller struct {
	ID           uuid.UUID
	Name         string
	RateLimitRPM int
}

type callerKey struct{}

func GetCallerFromContext(ctx context.Context) (*Caller, bool) {
	c, ok := ctx.Value(callerKey{}).(*Caller)
	return c, ok
}
