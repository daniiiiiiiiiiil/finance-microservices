include .env
export

db-up:
	docker compose -f ../docker-compose.yml up -d saga-postgres

db-down:
	docker compose -f ../docker-compose.yml down saga-postgres

db-restart:
	docker compose -f ../docker-compose.yml restart saga-postgres

db-logs:
	docker compose -f ../docker-compose.yml logs -f saga-postgres

db-cleanup:
	@read -p "Очистить все данные PostgreSQL? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose -f ../docker-compose.yml down saga-postgres && \
	  rm -rf ../out/pgdata/saga && \
	  echo "Данные PostgreSQL очищены"; \
	else \
	  echo "Очистка отменена"; \
	fi

migrate-up:
	@docker compose -f ../docker-compose.yml run --rm saga-postgres-migrate \
		-path /migrations \
		-database "postgres://${SAGA_POSTGRES_USER}:${SAGA_POSTGRES_PASSWORD}@saga-postgres:5432/${SAGA_POSTGRES_DB}?sslmode=disable" \
		up

migrate-down:
	@docker compose -f ../docker-compose.yml run --rm saga-postgres-migrate \
		-path /migrations \
		-database "postgres://${SAGA_POSTGRES_USER}:${SAGA_POSTGRES_PASSWORD}@saga-postgres:5432/${SAGA_POSTGRES_DB}?sslmode=disable" \
		down $(N)

migrate-down-all:
	@docker compose -f ../docker-compose.yml run --rm saga-postgres-migrate \
		-path /migrations \
		-database "postgres://${SAGA_POSTGRES_USER}:${SAGA_POSTGRES_PASSWORD}@saga-postgres:5432/${SAGA_POSTGRES_DB}?sslmode=disable" \
		down -all

migrate-force:
	@docker compose -f ../docker-compose.yml run --rm saga-postgres-migrate \
		-path /migrations \
		-database "postgres://${SAGA_POSTGRES_USER}:${SAGA_POSTGRES_PASSWORD}@saga-postgres:5432/${SAGA_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version:
	@docker compose -f ../docker-compose.yml run --rm saga-postgres-migrate \
		-path /migrations \
		-database "postgres://${SAGA_POSTGRES_USER}:${SAGA_POSTGRES_PASSWORD}@saga-postgres:5432/${SAGA_POSTGRES_DB}?sslmode=disable" \
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
	docker compose -f ../docker-compose.yml up -d saga

down:
	docker compose -f ../docker-compose.yml down saga

restart:
	docker compose -f ../docker-compose.yml restart saga

build:
	docker compose -f ../docker-compose.yml up -d --build saga

logs:
	docker compose -f ../docker-compose.yml logs -f saga

ps:
	docker compose -f ../docker-compose.yml ps saga

run:
	go run cmd/saga/main.go

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
	go build -o bin/saga-service cmd/saga/main.go

clean:
	rm -rf bin/
	rm -rf out/
	go clean


env-up:
	docker compose -f ../docker-compose.yml up -d saga-postgres

env-down:
	docker compose -f ../docker-compose.yml down saga-postgres

env-restart:
	docker compose -f ../docker-compose.yml restart saga-postgres