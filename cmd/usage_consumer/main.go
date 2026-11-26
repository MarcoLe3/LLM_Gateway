package main

import (
	"context"
	"llm_gateway/internal/config"
	"llm_gateway/internal/db"
	"llm_gateway/internal/events"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	store := events.NewEventStore(pool)
	consumer := events.NewKafkaConsumer(strings.Split(cfg.KafkaBrokers, ","), "usage_writer", store)
	slog.Info("starting usage event consumer")
	if err := consumer.Run(ctx); err != nil {
		log.Fatalf("consumer stopped: %v", err)
	}
	slog.Info("ending usage event consumer")
}
