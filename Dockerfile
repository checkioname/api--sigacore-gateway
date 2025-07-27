# --- Estágio 1: Builder ---
# Usamos uma imagem Go com Alpine para um build mais leve.
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte do projeto
COPY . .

# Argumento para decidir qual binário construir
# 'gateway' ou 'auth-service'
ARG cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main.go ./cmd/${cmd}

# --- Estágio 2: Final ---
FROM alpine:latest

WORKDIR /app

# Copia apenas o binário compilado do estágio de build.
COPY --from=builder /app/main /app/main

EXPOSE 8080

CMD ["/app/main"]