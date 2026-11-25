package main

import (
	"context"
	"llm_gateway/internal/api"
	"llm_gateway/internal/auth"
	"llm_gateway/internal/cache"
	"llm_gateway/internal/config"
	"llm_gateway/internal/db"
	"llm_gateway/internal/limiter"
	"llm_gateway/internal/provider"
	"llm_gateway/internal/queue"
	"llm_gateway/internal/router"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database pool: %v", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	cache := cache.NewCache(rdb, 10*time.Minute)

	rq, err := queue.NewRabbitQueue(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}

	defer rq.Close()

	authSvc := auth.NewAuth(*auth.NewAuthStore(pool), nil)
	lim := limiter.NewLimiter(rdb)
	rtr := router.New(
		[]provider.Provider{
			provider.NewClaude(cfg.AnthropicKey, "your-claude-model"),
			provider.NewOpenAI(cfg.OpenAIKey, "your-openai-model"),
		},
		cfg.RouterAttempts, 200*time.Millisecond, cfg.ProviderTimeout,
	)

	h := api.NewHandler(authSvc, lim, rtr, cache, pub, checks)

	go func() {

	}

	<- ctx.Done()
	slog.Info("shutdown")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = pub.Close()
}
