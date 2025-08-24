# Dockerfile para Gateway
FROM golang:1.22-alpine AS builder

# Instalar dependências
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copiar go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o gateway cmd/main.go

# Runtime image
FROM alpine:3.18

RUN apk --no-cache add ca-certificates wget
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copiar binário
COPY --from=builder /app/gateway .
RUN chown appuser:appgroup gateway

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./gateway"]