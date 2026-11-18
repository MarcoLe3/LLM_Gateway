package auth

import "uuid"

type Caller struct {
	ID			uuid.UUID
	Name		string
	RateLimitRPM int
}