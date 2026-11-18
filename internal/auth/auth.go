package auth

import (
	"context"
	"errors"

	"llm_gateway/internal/cache"
)

type Auth struct {
	store AuthStore
	cache cache.Cache
}

var ErrInvalidKey = errors.New("Invalid api key")

func NewAuth(store AuthStore, cache cache.Cache) *Auth {
	return &Auth{
		store: store,
		cache: cache,
	}
}

func (a *Auth) Authenticate(ctx context.Context, raw string) (*Caller, error) {
	hash := HashKey(raw)

	
}