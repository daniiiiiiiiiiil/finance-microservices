include .env
export

env-up:
	docker compose up -d finance-postgres finance-redis

env-down:
	docker compose down finance-postgres finance-redis

env-restart:
	docker compose restart finance-postgres finance-redis

env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных.[y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down finance-postgres finance-redis && \
	  rm -rf out/pgdata && \
	  rm -rf out/redisdata && \
	  echo "Файлы окружения очищены"; \
	else \
	  echo "Очистка окружения отменена"; \
	fi

env-port-forward:
	docker compose up -d port-forwarder

env-port-close:
	docker compose down port-forwarder


redis-up:
	docker compose up -d finance-redis

redis-down:
	docker compose down finance-redis

redis-restart:
	docker compose restart finance-redis

redis-logs:
	docker compose logs finance-redis

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
	docker compose up -d zookeeper kafka kafka-ui

kafka-down:
	docker compose down zookeeper kafka kafka-ui

kafka-restart:
	docker compose restart zookeeper kafka kafka-ui

kafka-logs:
	docker compose logs -f zookeeper kafka kafka-ui

kafka-ui:
	@echo "Kafka UI доступен по адресу: http://localhost:8080"
	@xdg-open http://localhost:8080 || open http://localhost:8080 || echo "Откройте браузер и перейдите по ссылке"

kafka-shell:
	docker compose exec -it kafka bash

kafka-create-topic:
	@docker compose exec kafka kafka-topics --create --topic $(TOPIC) --bootstrap-server kafka:9092 --partitions 1 --replication-factor 1

kafka-list-topics:
	@docker compose exec kafka kafka-topics --list --bootstrap-server kafka:9092

migrate-create:
	@docker compose run --rm finance-postgres-migrate \
	create \
	-ext sql \
	-dir /migrations \
	-seq $(seq)

migrate-up-auth:
	@docker compose run --rm auth-postgres-migrate \
		-path /migrations \
		-database postgres://${AUTH_POSTGRES_USER}:${AUTH_POSTGRES_PASSWORD}@auth-postgres:5432/${AUTH_POSTGRES_DB}?sslmode=disable \
		up

migrate-down-auth:
	@docker compose run --rm auth-postgres-migrate \
		-path /migrations \
		-database postgres://${AUTH_POSTGRES_USER}:${AUTH_POSTGRES_PASSWORD}@auth-postgres:5432/${AUTH_POSTGRES_DB}?sslmode=disable \
		down $(N)

migrate-down-all-auth:
	@docker compose run --rm auth-postgres-migrate \
		-path /migrations \
		-database postgres://${AUTH_POSTGRES_USER}:${AUTH_POSTGRES_PASSWORD}@auth-postgres:5432/${AUTH_POSTGRES_DB}?sslmode=disable \
		down -all

migrate-force-auth:
	@docker compose run --rm auth-postgres-migrate \
		-path /migrations \
		-database "postgres://${AUTH_POSTGRES_USER}:${AUTH_POSTGRES_PASSWORD}@auth-postgres:5432/${AUTH_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version-auth:
	@docker compose run --rm auth-postgres-migrate \
		-path /migrations \
		-database "postgres://${AUTH_POSTGRES_USER}:${AUTH_POSTGRES_PASSWORD}@auth-postgres:5432/${AUTH_POSTGRES_DB}?sslmode=disable" \
		version

migrate-up-users:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable \
		up

migrate-down-users:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable \
		down $(N)

migrate-down-all-users:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable \
		down -all

migrate-force-users:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version-users:
	@docker compose run --rm users-postgres-migrate \
		-path /migrations \
		-database "postgres://${USERS_POSTGRES_USER}:${USERS_POSTGRES_PASSWORD}@users-postgres:5432/${USERS_POSTGRES_DB}?sslmode=disable" \
		version

migrate-up-finance:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable \
		up

migrate-down-finance:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable \
		down $(N)

migrate-down-all-finance:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable \
		down -all

migrate-force-finance:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version-finance:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${FINANCE_POSTGRES_USER}:${FINANCE_POSTGRES_PASSWORD}@finance-postgres:5432/${FINANCE_POSTGRES_DB}?sslmode=disable" \
		version

migrate-up-all:
	@echo "Applying Auth migrations..."
	@$(MAKE) migrate-up-auth
	@echo "Applying Users migrations..."
	@$(MAKE) migrate-up-users
	@echo "Applying Finance migrations..."
	@$(MAKE) migrate-up-finance
	@echo "All migrations applied!"

migrate-down-all:
	@echo "Rolling back Finance migrations..."
	@$(MAKE) migrate-down-finance
	@echo "Rolling back Users migrations..."
	@$(MAKE) migrate-down-users
	@echo "Rolling back Auth migrations..."
	@$(MAKE) migrate-down-auth
	@echo "All migrations rolled back!"


finance-deploy:
	docker compose up -d --build finance

finance-undeploy:
	docker compose down finance

finance-logs:
	docker compose logs -f finance

finance-restart:
	docker compose restart finance


swagger-gen:
	docker compose run --rm swagger \
		init \
		-g cmd/gateway/main.go \
		-o docs \
		--parseInternal \
		--parseDependency \
		--parseVendor

up:
	docker compose up -d

down:
	docker compose down

build:
	docker-compose up -d --build

logs:
	docker compose logs -f

ps:
	docker compose ps

cleanup:
	@read -p "Очистить все данные (PostgreSQL + Redis)? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down && \
	  rm -rf out/pgdata && \
	  rm -rf out/redisdata && \
	  echo "Все данные очищены"; \
	else \
	  echo "Очистка отменена"; \
	fi


auth-up:
	docker compose up -d --build auth

users-up:
	docker compose up -d --build users

admin-up:
	docker compose up -d --build admin

finance-up:
	docker compose up -d --build finance

gateway-up:
	docker compose up -d --build gateway

services-up:
	docker compose up -d --build auth users admin finance gateway

services-down:
	docker compose down auth users admin finance gateway

services-restart:
	docker compose restart auth users admin finance gateway

auth-logs:
	docker compose logs -f auth

users-logs:
	docker compose logs -f users

admin-logs:
	docker compose logs -f admin

finance-logs:
	docker compose logs -f finance

gateway-logs:
	docker compose logs -f gateway

services-logs:
	docker compose logs -f auth users admin finance gateway

services-ps:
	docker compose ps auth users admin finance gateway

run-auth:
	@set ADDRESS=:5051 && go run cmd/auth/main.go

run-users:
	@set ADDRESS=:5052 && go run cmd/users/main.go

run-admin:
	@set ADDRESS=:5053 && go run cmd/admin/main.go

run-finance:
	@set ADDRESS=:5050 && go run cmd/finance/main.go

run-gateway:
	@set ADDRESS=:8080 && go run cmd/gateway/main.go


proto-gen-users:
	@echo "Generating users proto..."
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

proto-gen-auth:
	@echo "Generating auth proto..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		proto/auth/auth.proto

proto-gen-finance:
	@echo "Generating finance proto..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		proto/finance/finance.proto

proto-gen-admin:
	@echo "Generating admin proto..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		proto/admin/admin.proto

proto-gen-all:
	@echo "Generating all proto files..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		proto/auth/auth.proto \
		proto/users/users.proto \
		proto/finance/finance.proto \
		proto/admin/admin.proto
	@echo "All proto files generated!"