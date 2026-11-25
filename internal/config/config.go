package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	DatabaseURL  string
	RedisAddr    string
	RabbitURL    string
	KafkaBrokers string
	AnthropicKey string
	OpenAIKey    string

	Workers                    int
	MaxAttempts                int
	RouterAttempts             int
	ProviderTimeout            time.Duration
	RateLimitAllowOnRedisError bool
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  mustEnv("DATABASE_URL"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RabbitURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		AnthropicKey: os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIKey:    os.Getenv("OPENAI_API_KEY"),

		Workers:         getInt("WORKERS", 3),
		MaxAttempts:     getInt("MAX_ATTEMPTS", 3),
		RouterAttempts:  getInt("ROUTER_ATTEMPTS", 2),
		ProviderTimeout: time.Duration(getInt("PROVIDER_TIMEOUT_SECONDS", 30)) * time.Second,

		RateLimitAllowOnRedisError: getEnv("RATELIMIT_ALLOW_ON_REDIS_ERROR", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err != nil {
			return n
		}
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var %s", key)
	}
	return v
}
