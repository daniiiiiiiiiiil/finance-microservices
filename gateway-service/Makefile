include .env
export

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
		-I./proto \
		proto/auth/v1/auth.proto \
		proto/users/v1/users.proto \
		proto/finance/v1/finance.proto \
		proto/admin/v1/admin.proto \
		proto/currency/v1/currency.proto
	@echo "All proto files generated!"

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
		-I./proto \
		proto/auth/v1/auth.proto

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
		proto/users/v1/users.proto

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
		-I./proto \
		proto/finance/v1/finance.proto

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
		-I./proto \
		proto/admin/v1/admin.proto

proto-gen-currency:
	@echo "Generating currency proto..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		-I. \
		-I./proto \
		proto/currency/v1/currency.proto


swagger-gen:
	@echo "Generating Swagger from proto files..."
	protoc \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--openapiv2_out=docs \
		--openapiv2_opt=allow_merge=true,merge_file_name=swagger \
		-I. \
		-I./proto \
		proto/auth/v1/auth.proto \
		proto/users/v1/users.proto \
		proto/finance/v1/finance.proto \
		proto/admin/v1/admin.proto \
		proto/currency/v1/currency.proto
	@echo "Swagger merged into docs/swagger.json"

swagger-gen-auth:
	@echo "Generating auth Swagger..."
	protoc \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs \
		--openapiv2_opt=allow_merge=true,merge_file_name=auth \
		-I. \
		-I./proto \
		proto/auth/v1/auth.proto

swagger-gen-users:
	@echo "Generating users Swagger..."
	protoc \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs \
		--openapiv2_opt=allow_merge=true,merge_file_name=users \
		-I. \
		-I./proto \
		proto/users/v1/users.proto

swagger-gen-finance:
	@echo "Generating finance Swagger..."
	protoc \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs \
		--openapiv2_opt=allow_merge=true,merge_file_name=finance \
		-I. \
		-I./proto \
		proto/finance/v1/finance.proto

swagger-gen-admin:
	@echo "Generating admin Swagger..."
	protoc \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs \
		--openapiv2_opt=allow_merge=true,merge_file_name=admin \
		-I. \
		-I./proto \
		proto/admin/v1/admin.proto

swagger-gen-currency:
	@echo "Generating currency Swagger..."
	protoc \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs \
		--openapiv2_opt=allow_merge=true,merge_file_name=currency \
		-I. \
		-I./proto \
		proto/currency/v1/currency.proto


gateway-up:
	docker compose up -d --build gateway

gateway-down:
	docker compose down gateway

gateway-restart:
	docker compose restart gateway

gateway-logs:
	docker compose logs -f gateway

gateway-build:
	docker compose build gateway


run:
	@set ADDRESS=:8081 && go run cmd/gateway/main.go


test:
	go test ./... -v


build:
	go build ./...


clean-swagger:
	rm -rf docs/swagger.json docs/auth.json docs/users.json docs/finance.json docs/admin.json docs/currency.json

clean-proto:
	rm -rf proto/auth/gen proto/users/gen proto/finance/gen proto/admin/gen proto/currency/gen

clean:
	@$(MAKE) clean-swagger
	@$(MAKE) clean-proto
	@echo "All generated files cleaned!"