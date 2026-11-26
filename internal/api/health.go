package api

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DepCheck struct {
	Name     string
	Critical bool
	Ping     func(ctx context.Context) error
}

func PostgresCheck(pool *pgxpool.Pool) DepCheck {
	return DepCheck{Name: "postgres", Critical: true, Ping: pool.Ping}
}

func RedisCheck(rdb *redis.Client) DepCheck {
	return DepCheck{
		Name:     "redis",
		Critical: false,
		Ping: func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
	}
}
