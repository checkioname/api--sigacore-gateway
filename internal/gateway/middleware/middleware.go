package middleware

import (
	"api--sigacore-gateway/internal/token"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/time/rate"
)

type contextKey string

const authPayloadKey = contextKey("authorization_payload")

// IPWhitelist verifica se o IP da requisição está na lista de permissão.
func IPWhitelist(allowedIPs ...string) func(http.Handler) http.Handler {
	whitelist := make(map[string]struct{})
	for _, ip := range allowedIPs {
		whitelist[ip] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			}

			if _, found := whitelist[ip]; !found {
				log.Printf("Blocked: IP %s is not in whitelist", ip)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter cria um middleware de limite de requisições.
func RateLimiter(r rate.Limit, b int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(r, b)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware verifica o token de autenticação Paseto para o gateway.
func AuthMiddleware(tokenMaker token.Maker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header is required", http.StatusUnauthorized)
				return
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			authType := strings.ToLower(fields[0])
			if authType != "bearer" {
				http.Error(w, fmt.Sprintf("unsupported authorization type: %s", authType), http.StatusUnauthorized)
				return
			}

			accessToken := fields[1]
			payload, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Adiciona o payload ao contexto da requisição para uso futuro
			ctx := context.WithValue(r.Context(), authPayloadKey, payload)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
