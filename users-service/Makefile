include .env
export

migrate-create-local:
	migrate create -ext sql -dir ./migrations -seq $(seq)

migrate-create:
	@docker compose run --rm users-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq $(seq)

migrate-up:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		up

migrate-down:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		down $(N)

migrate-down-all:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		down -all

migrate-force:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		version

up:
	docker compose up -d users

down:
	docker compose down users

restart:
	docker compose restart users

build:
	docker compose up -d --build users

logs:
	docker compose logs -f users

ps:
	docker compose ps users

run:
	go run cmd/main.go

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy


proto-gen:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		-I./proto \
		proto/users/users.proto

