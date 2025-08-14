.PHONY: build clean test run-auth run-gateway run-combined docker-build docker-run migrate-up migrate-down sqlc generate-key check-security

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

# Security and Configuration
generate-key:
	@echo "🔐 Gerando chave segura..."
	@mkdir -p bin
	@go build -o bin/generate-key scripts/generate-key.go
	@./bin/generate-key

generate-key-multiple:
	@echo "🔐 Gerando múltiplas chaves para rotação..."
	@mkdir -p bin
	@go build -o bin/generate-key scripts/generate-key.go
	@./bin/generate-key --multiple

check-security:
	@echo "🔍 Verificando configurações de segurança..."
	@echo "Verificando se chaves padrão estão sendo usadas..."
	@if grep -q "12345678901234567890123456789012" app.env 2>/dev/null; then \
		echo "❌ ERRO: Chave padrão detectada em app.env"; \
		echo "   Execute: make generate-key"; \
		exit 1; \
	fi
	@if grep -q "DEV_KEY_NOT_FOR_PRODUCTION" app.env 2>/dev/null; then \
		echo "⚠️  AVISO: Chave de desenvolvimento detectada"; \
		echo "   Para produção, execute: make generate-key"; \
	fi
	@if [ "$(ENVIRONMENT)" = "production" ] && grep -q "localhost" app.env 2>/dev/null; then \
		echo "❌ ERRO: Configurações de localhost em produção"; \
		exit 1; \
	fi
	@echo "✅ Verificação de segurança concluída"

check-prod-config:
	@echo "🔍 Verificando configuração para produção..."
	@if [ -z "$(TOKEN_SYMMETRIC_KEY)" ]; then \
		echo "❌ ERRO: TOKEN_SYMMETRIC_KEY não definida"; \
		exit 1; \
	fi
	@if [ "$(TOKEN_SYMMETRIC_KEY)" = "DEV_KEY_NOT_FOR_PRODUCTION_USE_32" ]; then \
		echo "❌ ERRO: Usando chave de desenvolvimento em produção"; \
		exit 1; \
	fi
	@if [ -z "$(ENVIRONMENT)" ] || [ "$(ENVIRONMENT)" != "production" ]; then \
		echo "❌ ERRO: ENVIRONMENT deve ser 'production'"; \
		exit 1; \
	fi
	@echo "✅ Configuração de produção válida"

setup-prod-example:
	@echo "📋 Criando exemplo de configuração de produção..."
	@mkdir -p deployment
	@if [ ! -f deployment/production.env.example ]; then \
		echo "❌ Arquivo deployment/production.env.example não encontrado"; \
		exit 1; \
	fi
	@echo "✅ Template de produção disponível em: deployment/production.env.example"
	@echo "   Copie e configure com valores reais antes do deploy"

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

# Lint e verificações de código
lint:
	golangci-lint run

# Security scan
security-scan:
	@echo "🔍 Executando scan de segurança..."
	@echo "Verificando dependências com vulnerabilidades conhecidas..."
	@go list -json -m all | nancy sleuth
	@echo "Verificando secrets no código..."
	@git secrets --scan || echo "⚠️  git-secrets não instalado. Instale com: brew install git-secrets"

# Setup completo para desenvolvimento
setup-dev: install-tools generate-key
	@echo "🚀 Configuração de desenvolvimento concluída!"
	@echo "   1. Configure as variáveis no app.env"
	@echo "   2. Execute: make dev"

# Help
help:
	@echo "Comandos disponíveis:"
	@echo ""
	@echo "BUILD:"
	@echo "  build          - Compila todos os serviços"
	@echo "  build-auth     - Compila apenas o serviço de auth"
	@echo "  build-gateway  - Compila apenas o gateway"
	@echo ""
	@echo "DESENVOLVIMENTO:"
	@echo "  dev            - Executa em modo desenvolvimento"
	@echo "  run-combined   - Executa ambos os serviços"
	@echo "  setup-dev      - Configuração inicial para desenvolvimento"
	@echo ""
	@echo "SEGURANÇA:"
	@echo "  generate-key   - Gera chave segura para tokens"
	@echo "  check-security - Verifica configurações de segurança"
	@echo "  security-scan  - Executa scan de vulnerabilidades"
	@echo ""
	@echo "PRODUÇÃO:"
	@echo "  check-prod-config  - Valida configuração de produção"
	@echo "  setup-prod-example - Cria template de configuração"
	@echo ""
	@echo "DATABASE:"
	@echo "  migrate-up     - Executa migrações"
	@echo "  migrate-down   - Reverte migrações"
	@echo "  sqlc           - Gera código a partir das queries SQL"
	@echo ""
	@echo "OUTROS:"
	@echo "  test           - Executa testes"
	@echo "  clean          - Remove arquivos compilados"
	@echo "  lint           - Executa linter"
	@echo "  docker-build   - Constrói imagem Docker"