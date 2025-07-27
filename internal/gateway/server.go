package gateway

import (
	"api--sigacore-gateway/internal/util"
	"log"
	"net/http"

	"api--sigacore-gateway/internal/gateway/router"
	"api--sigacore-gateway/internal/token"
)

type GatewayServer struct {
	config     util.Config
	router     *router.GatewayRouter
	tokenMaker token.Maker
}

func NewGatewayServer(cfg util.Config) (*GatewayServer, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, err
	}

	gatewayRouter, err := router.NewGatewayRouter(cfg, tokenMaker)
	if err != nil {
		return nil, err
	}

	return &GatewayServer{
		config:     cfg,
		router:     gatewayRouter,
		tokenMaker: tokenMaker,
	}, nil
}

func (s *GatewayServer) Start() error {
	log.Printf("🚀 Gateway configurado com sucesso!")
	log.Printf("📍 Serviços configurados:")
	log.Printf("   - Auth Service: %s", s.config.AuthServerAddress)
	log.Printf("   - User Service: %s", s.config.UserServiceAddress)
	log.Printf("   - Report Service: %s", s.config.ReportServiceAddress)
	log.Printf("   - Notification Service: %s", s.config.NotificationServiceAddress)
	log.Printf("🔒 IPs permitidos: %v", s.config.AllowedIPs)

	return http.ListenAndServe(s.config.GatewayServerAddress, s.router.Handler())
}
