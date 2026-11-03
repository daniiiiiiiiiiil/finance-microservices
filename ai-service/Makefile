include .env
export

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Service:"
	@echo "  service-up        Start ai-service via docker compose"
	@echo "  service-down      Stop ai-service"
	@echo "  service-restart   Restart ai-service"
	@echo "  service-logs      Tail ai-service logs"
	@echo "  service-build     Build ai-service image"
	@echo ""
	@echo "Local:"
	@echo "  run               Run locally (go run)"
	@echo "  build             go build ./..."
	@echo "  test              go test ./... -v"
	@echo "  test-cover        go test ./... -cover"
	@echo "  lint              go vet ./..."
	@echo ""
	@echo "Proto:"
	@echo "  proto-gen         Generate proto (buf)"
	@echo ""
	@echo "Env:"
	@echo "  env-up            Start ai dependencies (redis, postgres)"
	@echo "  env-down          Stop ai dependencies"
	@echo ""

service-up:
	docker compose up -d --build ai

service-down:
	docker compose down ai

service-restart:
	docker compose restart ai

service-logs:
	docker compose logs -f ai

service-build:
	docker compose build ai

run:
	go run cmd/ai/main.go

build:
	go build ./...

test:
	go test ./... -v

test-cover:
	go test ./... -cover

lint:
	go vet ./...

proto-gen:
	@echo "Generating ai proto..."
	buf generate
	@echo "Proto files generated!"

proto-gen-protoc:
	@echo "Generating ai proto via protoc..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		-I. \
		-I./proto \
		proto/ai/v1/ai.proto
	@echo "Proto files generated!"

env-up:
	docker compose up -d finance-redis ai-postgres

env-down:
	docker compose down finance-redis ai-postgres

env-restart:
	docker compose restart finance-redis ai-postgres

migrate-up:
	docker compose run --rm ai-postgres-migrate \
		-path /migrations \
		-database postgres://${AI_POSTGRES_USER}:${AI_POSTGRES_PASSWORD}@ai-postgres:5432/${AI_POSTGRES_DB}?sslmode=disable \
		up

migrate-down:
	docker compose run --rm ai-postgres-migrate \
		-path /migrations \
		-database postgres://${AI_POSTGRES_USER}:${AI_POSTGRES_PASSWORD}@ai-postgres:5432/${AI_POSTGRES_DB}?sslmode=disable \
		down $(N)

migrate-create:
	@echo "Creating migration: $(seq)"
	@docker run --rm \
		-v ./ai-service/migrations:/migrations \
		--entrypoint sh \
		migrate/migrate:latest \
		-c "migrate create -ext sql -dir /migrations -seq $(seq)"
	@echo "Migration created in ./ai-service/migrations/"

health:
	@echo "gRPC:    localhost:$(AI_GRPC_PORT)"
	@echo "Metrics: localhost:$(AI_METRICS_PORT)/metrics"
	@grpcurl -plaintext localhost:$(AI_GRPC_PORT) list || echo "grpcurl not installed"