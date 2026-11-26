.PHONY: help up down logs reset build run-gateway run-worker run-consumer \
	test test-up test-down bench-go

up:
	docker compose up -d --build

down:
	docker compose down

reset:
	docker compose down -v

logs:
	docker compose logs -f

run-gateway:
	go run ./cmd/gateway

run-worker:
	go run ./cmd/worker

test:
	go test ./...

test-up:
	docker compose -f docker-compose-test.yml up -d 

bench-go:
	go test -bench=. -benchmen ./internal/...

lint:
	go vet ./...