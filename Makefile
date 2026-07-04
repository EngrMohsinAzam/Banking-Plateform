.PHONY: help up down logs ps build run-api run-worker run-web test tidy migrate-up migrate-down test-integration test-integration-race test-load test-http vet docker-build

help:
	@echo "Targets:"
	@echo "  make up        - Start Postgres + Redis + Redpanda"
	@echo "  make down      - Stop infrastructure"
	@echo "  make logs      - Tail docker compose logs"
	@echo "  make ps        - Show container status"
	@echo "  make build     - Build api, worker, migrate binaries"
	@echo "  make run-api   - Run the HTTP API (dev)"
	@echo "  make run-worker - Run the background worker (dev)"
	@echo "  make run-web   - Run the Next.js frontend (dev)"
	@echo "  make migrate-up   - Apply database migrations"
	@echo "  make migrate-down - Roll back database migrations"
	@echo "  make test      - Run unit tests (no Docker required)"
	@echo "  make test-http - HTTP handler tests"
	@echo "  make test-integration - Integration tests (requires Docker)"
	@echo "  make test-integration-race - Integration tests with -race"
	@echo "  make test-load - Transfer throughput benchmark"
	@echo "  make vet       - go vet"
	@echo "  make docker-build - Build API and worker container images"
	@echo "  make tidy      - go mod tidy"

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/migrate ./cmd/migrate

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

run-web:
	cd web && npm run dev

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

test:
	go test ./...

test-http:
	go test ./internal/platform/http/...

vet:
	go vet ./...

test-integration:
	go test -tags=integration -v ./internal/ledger/adapters/postgres/... ./internal/transfer/app/... ./internal/settlement/app/... ./tests/integration/...

test-integration-race:
	go test -tags=integration -race -count=1 -v ./internal/ledger/adapters/postgres/... ./internal/transfer/app/... ./internal/settlement/app/...

test-load:
	go test -bench=BenchmarkTransferExecute -benchmem -count=1 ./tests/load/...

docker-build:
	docker build -f deploy/docker/api.Dockerfile -t banking-api:local .
	docker build -f deploy/docker/worker.Dockerfile -t banking-worker:local .

tidy:
	go mod tidy
