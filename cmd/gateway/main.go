package main

import (
	"context"
	"llm_gateway/internal/api"
	"llm_gateway/internal/auth"
	"llm_gateway/internal/cache"
	"llm_gateway/internal/config"
	"llm_gateway/internal/db"
	"llm_gateway/internal/events"
	"llm_gateway/internal/job"
	"llm_gateway/internal/limiter"
	"llm_gateway/internal/provider"
	"llm_gateway/internal/queue"
	"llm_gateway/internal/router"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	pub := events.NewKafkaProducer(strings.Split(cfg.KafkaBrokers, ","))
	defer pub.Close()

	authSvc := auth.NewAuth(*auth.NewAuthStore(pool))
	lim := limiter.NewLimiter(rdb, cfg.RateLimitAllowOnRedisError)
	rtr := router.New(
		[]provider.Provider{
			provider.NewClaude(cfg.AnthropicKey, "your-claude-model"),
			provider.NewOpenAI(cfg.OpenAIKey, "your-openai-model"),
		},
		cfg.RouterAttempts, 200*time.Millisecond, cfg.ProviderTimeout,
	)

	jobSvc := job.NewService(job.NewJobStore(pool), rq, cfg.MaxAttempts)

	checks := []api.DepCheck{
		api.PostgresCheck(pool),
		api.RedisCheck(rdb),
	}

	h := api.NewHandler(authSvc, lim, jobSvc, rtr, cache, pub, checks)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		slog.Info("gateway listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}

	_ = pub.Close()
}
