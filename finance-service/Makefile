include .env
export

db-up:
	docker compose up -d finance-postgres

db-down:
	docker compose down finance-postgres

db-restart:
	docker compose restart finance-postgres

db-logs:
	docker compose logs -f finance-postgres

db-cleanup:
	@read -p "Очистить все данные PostgreSQL? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down finance-postgres && \
	  rm -rf out/pgdata/finance && \
	  echo "Данные PostgreSQL очищены"; \
	else \
	  echo "Очистка отменена"; \
	fi


redis-up:
	docker compose up -d finance-redis

redis-down:
	docker compose down finance-redis

redis-restart:
	docker compose restart finance-redis

redis-logs:
	docker compose logs -f finance-redis

redis-cli:
	docker compose exec finance-redis redis-cli

redis-cleanup:
	@read -p "Очистить все данные Redis? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down finance-redis && \
	  rm -rf out/redisdata && \
	  echo "Данные Redis очищены"; \
	else \
	  echo "Очистка Redis отменена"; \
	fi


kafka-up:
	docker compose up -d kafka kafka-ui

kafka-down:
	docker compose down kafka kafka-ui

kafka-logs:
	docker compose logs -f kafka kafka-ui

kafka-ui:
	@echo "Kafka UI доступен по адресу: http://localhost:8080"
	@xdg-open http://localhost:8080 || open http://localhost:8080 || echo "Откройте браузер и перейдите по ссылке"

kafka-list-topics:
	@docker compose exec kafka kafka-topics --list --bootstrap-server kafka:9092

kafka-create-topic:
	@docker compose exec kafka kafka-topics --create --topic $(TOPIC) --bootstrap-server kafka:9092 --partitions 1 --replication-factor 1


migrate-up:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable \
		up

migrate-down:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable \
		down $(N)

migrate-down-all:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable \
		down -all

migrate-force:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable" \
		version

migrate-create:
	@echo "Creating migration: $(seq)"
	@docker run --rm \
		-v ./migrations:/migrations \
		--entrypoint sh \
		migrate/migrate:latest \
		-c "migrate create -ext sql -dir /migrations -seq $(seq)"
	@echo "Migration created in ./migrations/"


s3-up:
	docker compose up -d minio

s3-down:
	docker compose down minio

s3-logs:
	docker compose logs -f minio

s3-bucket-create:
	docker exec finance-minio mc mb --ignore-existing local/finance-files

s3-bucket-list:
	docker exec finance-minio mc ls local/

s3-console:
	@echo "MinIO Console доступен по адресу: http://localhost:9001"
	@xdg-open http://localhost:9001 || open http://localhost:9001 || echo "Откройте браузер и перейдите по ссылке"


service-up:
	docker compose up -d --build finance

service-down:
	docker compose down finance

service-restart:
	docker compose restart finance

service-logs:
	docker compose logs -f finance

service-build:
	docker compose build finance

proto-gen:
	@echo "Generating finance proto..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		-I./proto \
		proto/finance/v1/finance.proto
	@echo "Proto files generated!"


run:
	@set ADDRESS=:50053 && go run cmd/finance/main.go


test:
	go test ./... -v

test-cover:
	go test ./... -cover


build:
	go build ./...


env-up:
	docker compose up -d finance-postgres finance-redis minio

env-down:
	docker compose down finance-postgres finance-redis minio

env-restart:
	docker compose restart finance-postgres finance-redis minio