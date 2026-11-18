-include .env
export

.PHONY: run build test test-race vet proto-gen migrate-up migrate-down

MIGRATE := migrate -path migrations

run:
	go run ./cmd/telegram/main.go

build:
	go build ./...

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

proto-gen:
	protoc -I proto \
		--go_out=. \
		--go_opt=module=github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=module=github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service \
		proto/telegram/v1/telegram.proto

migrate-up:
	$(MIGRATE) -database "$(TELEGRAM_DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -database "$(TELEGRAM_DATABASE_URL)" down

migrate-up-docker:
	docker compose run --rm telegram-migrate