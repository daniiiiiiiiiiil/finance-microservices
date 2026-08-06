include .env
export

env-up:
	docker compose up -d finance-postgres

env-down:
	docker compose down finance-postgres

env-restart:
	docker compose restart

env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных.[y/n]": ans; \
	if ["$$ans" = "y"]; then \
	  docker compose down finance-postgres && \
	  rm -rf out/pgdata && \
	  echo "Файлы окружения очищены"; \
	else \
	  echo "Очистка окружения отменена";\
	fi

env-port-forward:
	docker compose up -d port-forwarder

env-port-close:
	docker compose down port-forwarder

migrate-create:
	@docker compose run --rm finance-postgres-migrate \
	create \
	-ext sql \
	-dir /migrations \
	-seq $(seq)

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

migrate-force:
	docker compose run --rm finance-postgres-migrate \
		-path /migrations \
		-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@finance-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
		force $(version)

finance-deploy:
	docker compose up -d --build finance

finance-undeploy:
	docker compose down finance

swagger-gen:
	docker compose run --rm swagger \
		init \
		-g cmd/finance/main.go \
		-o docs \
		--parseInternal \
		--parseDependency \