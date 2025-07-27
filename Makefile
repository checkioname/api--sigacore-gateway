.PHONY: build clean test run-auth run-gateway run-combined docker-build docker-run migrate-up migrate-down sqlc

# Configurações
DB_URL=postgresql://root:secret@localhost:5432/sigacore?sslmode=disable

# Build targets
build-auth:
	go build -o bin/auth-service cmd/auth-service/main.go

build-gateway:
	go build -o bin/gateway cmd/gateway/main.go

build-combined:
	go build -o bin/combined cmd/combined/main.go

build: build-auth build-gateway build-combined

# Run targets
run-auth:
	go run cmd/auth-service/main.go

run-gateway:
	go run cmd/gateway/main.go

run-combined:
	go run cmd/combined/main.go

# Development
dev: run-combined

# Test
test:
	go test -v -cover ./...

# Database migrations
migrate-up:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migrate-create:
	migrate create -ext sql -dir db/migration -seq $(name)

# SQLC
sqlc:
	sqlc generate

# Docker
docker-build:
	docker build -t sigacore-gateway .

docker-run:
	docker run --name sigacore-gateway -p 8081:8081 sigacore-gateway

# Clean
clean:
	rm -rf bin/

# Proto (se você usar gRPC no futuro)
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# Instalação de dependências de desenvolvimento
install-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

# Help
help:
	@echo "Comandos disponíveis:"
	@echo "  build-auth     - Compila o serviço de autenticação"
	@echo "  build-gateway  - Compila o gateway"
	@echo "  build-combined - Compila a versão combinada"
	@echo "  build          - Compila todos os serviços"
	@echo "  run-auth       - Executa apenas o serviço de autenticação"
	@echo "  run-gateway    - Executa apenas o gateway"
	@echo "  run-combined   - Executa ambos os serviços simultaneamente"
	@echo "  dev            - Alias para run-combined"
	@echo "  test           - Executa os testes"
	@echo "  migrate-up     - Executa migrações do banco"
	@echo "  migrate-down   - Reverte migrações do banco"
	@echo "  sqlc           - Gera código Go a partir das queries SQL"
	@echo "  docker-build   - Constrói imagem Docker"
	@echo "  docker-run     - Executa container Docker"
	@echo "  clean          - Remove arquivos compilados"
	@echo "  install-tools  - Instala ferramentas de desenvolvimento"