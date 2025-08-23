package gateway

import (
	"log"

	"github.com/gin-gonic/gin"

	"api--sigacore-gateway/internal/gateway/router"
	"api--sigacore-gateway/internal/token"
	"api--sigacore-gateway/internal/util"
)

type GatewayServer struct {
	config     util.Config
	router     *gin.Engine
	tokenMaker token.Maker
}

func NewGatewayServer(cfg util.Config) (*GatewayServer, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, err
	}

	return &GatewayServer{
		config:     cfg,
		router:     router.SetupGatewayRoutes(cfg),
		tokenMaker: tokenMaker,
	}, nil
}

func (s *GatewayServer) Start() error {
	log.Printf("🚀 Gateway configurado com sucesso!")
	log.Printf("📍 Serviços configurados:")
	log.Printf("   - Auth Service: %s", s.config.AuthServerAddress)
	log.Printf("   - User Service: %s", s.config.UserServiceAddress)
	log.Printf("   - Report Service: %s", s.config.DocServiceAddress)
	log.Printf("🔒 IPs permitidos: %v", s.config.AllowedIPs)

	return s.router.Run(s.config.GatewayServerAddress)
}
