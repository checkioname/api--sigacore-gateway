package router

import (
	"api--sigacore-gateway/internal/token"
	"api--sigacore-gateway/internal/util"
	"net/http"

	"api--sigacore-gateway/internal/gateway/middleware"
	"api--sigacore-gateway/internal/gateway/proxy"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"
)

type GatewayRouter struct {
	config     util.Config
	tokenMaker token.Maker
	router     *chi.Mux
}

func NewGatewayRouter(cfg util.Config, tokenMaker token.Maker) (*GatewayRouter, error) {
	r := &GatewayRouter{
		config:     cfg,
		tokenMaker: tokenMaker,
		router:     chi.NewRouter(),
	}

	if err := r.setupMiddlewares(); err != nil {
		return nil, err
	}

	if err := r.setupRoutes(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *GatewayRouter) setupMiddlewares() error {
	// Middleware global de whitelist de IPs
	ipWhitelistMiddleware := middleware.IPWhitelist(r.config.AllowedIPs...)

	// Middleware global de rate limiting
	rateLimiterMiddleware := middleware.RateLimiter(rate.Limit(5), 10) // 5 req/s, burst de 10

	r.router.Use(ipWhitelistMiddleware)
	r.router.Use(rateLimiterMiddleware)

	return nil
}

func (r *GatewayRouter) setupRoutes() error {
	// Criar proxies para os serviços
	authServiceProxy, err := proxy.NewReverseProxy(r.config.AuthServerAddress)
	if err != nil {
		return err
	}

	userServiceProxy, err := proxy.NewReverseProxy(r.config.UserServiceAddress)
	if err != nil {
		return err
	}

	reportServiceProxy, err := proxy.NewReverseProxy(r.config.ReportServiceAddress)
	if err != nil {
		return err
	}

	notificationServiceProxy, err := proxy.NewReverseProxy(r.config.NotificationServiceAddress)
	if err != nil {
		return err
	}

	// Middleware de autenticação
	authMiddleware := middleware.AuthMiddleware(r.tokenMaker)

	// Rotas públicas para o serviço de autenticação
	r.router.Mount("/auth", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Remove o prefixo /auth da URL antes de enviar para o serviço
		req.URL.Path = req.URL.Path[len("/auth"):]

		// Apenas login, criação de usuário e refresh de token são públicos
		if req.URL.Path == "/users/login" || req.URL.Path == "/users" || req.URL.Path == "/token/renew" {
			authServiceProxy.ServeHTTP(w, req)
			return
		}
		// Para qualquer outra rota dentro de /auth/, aplicamos a autenticação
		authMiddleware(authServiceProxy).ServeHTTP(w, req)
	}))

	// Grupo de rotas protegidas que precisam de autenticação
	r.router.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		// Serviço de usuários (gerenciamento de perfis, etc.)
		r.Mount("/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = req.URL.Path[len("/users"):]
			userServiceProxy.ServeHTTP(w, req)
		}))

		// Serviço de relatórios
		r.Mount("/reports", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = req.URL.Path[len("/reports"):]
			reportServiceProxy.ServeHTTP(w, req)
		}))

		// Serviço de notificações
		r.Mount("/notifications", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = req.URL.Path[len("/notifications"):]
			notificationServiceProxy.ServeHTTP(w, req)
		}))
	})

	// Health check
	r.router.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gateway is healthy"))
	})

	return nil
}

func (r *GatewayRouter) Handler() http.Handler {
	return r.router
}
