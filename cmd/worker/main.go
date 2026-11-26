package consumer

import (
	"context"
	"llm_gateway/internal/config"
	"llm_gateway/internal/db"
	"llm_gateway/internal/events"
	"llm_gateway/internal/job"
	"llm_gateway/internal/provider"
	"llm_gateway/internal/queue"
	"llm_gateway/internal/router"
	"llm_gateway/internal/worker"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(cfg.DatabaseURL)
	defer pool.Close()

	rq, err := queue.NewRabbitQueue(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}

	defer rq.Close()

	pub := events.NewKafkaProducer(strings.Split(cfg.KafkaBrokers, ","))
	defer pub.Close()

	rtr := router.New(
		[]provider.Provider{
			provider.NewClaude(cfg.AnthropicKey, ""),
			provider.NewOpenAI(cfg.OpenAIKey, ""),
		},
		cfg.MaxAttempts, 200*time.Millisecond, cfg.ProviderTimeout,
	)

	jobSvc := job.NewService(job.NewJobStore(pool), rq, cfg.MaxAttempts)
	w := worker.NewWorker(jobSvc, rtr, rq, cfg)

	slog.Info("workers started", "workers", cfg.Workers)
	if err := w.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}
	slog.Info("worker stopped")
}
