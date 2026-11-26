package auth

import (
	"context"
	"errors"

	"llm_gateway/internal/cache"

	"github.com/jackc/pgx/v5"
)

type Auth struct {
	store AuthStore
	cache cache.Cache
}

var ErrInvalidKey = errors.New("Invalid api key")

func NewAuth(store AuthStore) *Auth {
	return &Auth{
		store: store,
	}
}

func (a *Auth) Authenticate(ctx context.Context, raw string) (*Caller, error) {
	hash := HashKey(raw)
	c, err := a.store.GetKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidKey
		}
		return nil, err
	}
	return c, nil
}
