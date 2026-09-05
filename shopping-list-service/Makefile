include .env
export

db-up:
	docker compose up -d shopping-postgres

db-down:
	docker compose down shopping-postgres

db-restart:
	docker compose restart shopping-postgres

db-logs:
	docker compose logs -f shopping-postgres

db-cleanup:
	@read -p "Очистить все данные PostgreSQL? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down shopping-postgres && \
	  rm -rf out/pgdata/shopping && \
	  echo "Данные PostgreSQL очищены"; \
	else \
	  echo "Очистка отменена"; \
	fi

redis-up:
	docker compose up -d shopping-redis

redis-down:
	docker compose down shopping-redis

redis-restart:
	docker compose restart shopping-redis

redis-logs:
	docker compose logs -f shopping-redis

redis-cli:
	docker compose exec shopping-redis redis-cli

redis-cleanup:
	@read -p "Очистить все данные Redis? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down shopping-redis && \
	  rm -rf out/redisdata/shopping && \
	  echo "Данные Redis очищены"; \
	else \
	  echo "Очистка Redis отменена"; \
	fi

migrate-up:
	@docker compose run --rm shopping-postgres-migrate \
		-path /migrations \
		-database "postgres://${SHOPPING_POSTGRES_USER}:${SHOPPING_POSTGRES_PASSWORD}@shopping-postgres:5432/${SHOPPING_POSTGRES_DB}?sslmode=disable" \
		up

migrate-down:
	@docker compose run --rm shopping-postgres-migrate \
		-path /migrations \
		-database "postgres://${SHOPPING_POSTGRES_USER}:${SHOPPING_POSTGRES_PASSWORD}@shopping-postgres:5432/${SHOPPING_POSTGRES_DB}?sslmode=disable" \
		down $(N)

migrate-down-all:
	@docker compose run --rm shopping-postgres-migrate \
		-path /migrations \
		-database "postgres://${SHOPPING_POSTGRES_USER}:${SHOPPING_POSTGRES_PASSWORD}@shopping-postgres:5432/${SHOPPING_POSTGRES_DB}?sslmode=disable" \
		down -all

migrate-force:
	@docker compose run --rm shopping-postgres-migrate \
		-path /migrations \
		-database "postgres://${SHOPPING_POSTGRES_USER}:${SHOPPING_POSTGRES_PASSWORD}@shopping-postgres:5432/${SHOPPING_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version:
	@docker compose run --rm shopping-postgres-migrate \
		-path /migrations \
		-database "postgres://${SHOPPING_POSTGRES_USER}:${SHOPPING_POSTGRES_PASSWORD}@shopping-postgres:5432/${SHOPPING_POSTGRES_DB}?sslmode=disable" \
		version

migrate-create-local:
	@migrate create -ext sql -dir ./migrations -seq $(seq)

migrate-create:
	@echo "Creating migration: $(seq)"
	@docker run --rm \
		-v ./migrations:/migrations \
		--entrypoint sh \
		migrate/migrate:latest \
		-c "migrate create -ext sql -dir /migrations -seq $(seq)"

up:
	docker compose up -d shopping

down:
	docker compose down shopping

restart:
	docker compose restart shopping

build:
	docker compose up -d --build shopping

logs:
	docker compose logs -f shopping

ps:
	docker compose ps shopping

run:
	go run cmd/main.go

test:
	go test ./... -race -cover

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

build-local:
	go build -o bin/shopping-service cmd/main.go

clean:
	rm -rf bin/
	rm -rf out/
	go clean

proto-gen:
	@echo "Generating shopping proto..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		-I./proto \
		proto/shopping/v1/shopping.proto

env-up:
	docker compose up -d shopping-postgres shopping-redis

env-down:
	docker compose down shopping-postgres shopping-redis

env-restart:
	docker compose restart shopping-postgres shopping-redis

