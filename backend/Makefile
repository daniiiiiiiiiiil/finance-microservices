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


migrate-create:
	@docker compose run --rm finance-postgres-migrate \
	create \
	-ext sql \
	-dir /migrations \
	-seq $(seq)

migrate-up:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		up

migrate-down:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		down $(N)

migrate-down-all:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		down -all

migrate-force:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-version:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
		version

migrate-action:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"


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
		-g cmd/finance/main.go \
		-o docs \
		--parseInternal \
		--parseDependency


up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

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