package api

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type depCheck struct {
	Name     string
	Critical bool
	Ping     func(ctx context.Context) error
}

func (h *Handler) SetDependencyChecks() {

}

func PostgresCheck(pool *pgxpool.Pool) depCheck {
	return depCheck{Name: "postgres", Critical: true, Ping: pool.Ping}
}

func RedisCheck(rdb *redis.Client) depCheck {
	return depCheck{
		Name:     "redis",
		Critical: false,
		Ping: func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
	}
}

func RabbitCheck(q *queue.RabbitQueue) depCheck {
	
}