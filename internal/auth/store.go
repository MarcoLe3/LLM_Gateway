package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthStore struct {
	DB *pgxpool.Pool
}

func NewAuthStore(pool *pgxpool.Pool) *AuthStore {
	return &AuthStore{
		DB: pool,
	}
}

func (s *AuthStore) GetKeyByHash(ctx context.Context, hash string) (*Caller, error) {
	var c Caller
	err := s.DB.QueryRow(ctx,
		`SELECT c.id, c.name, c.rate_limit_rpm 
		FROM api_keys k
		JOIN callers c ON k.caller_id = c.id
		WHERE k.key_hash = $1 AND revoked_at IS NULL`, hash).Scan(&c.ID, &c.Name, &c.RateLimitRPM)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("No caller")
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}
