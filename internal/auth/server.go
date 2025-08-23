package auth

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	"api--sigacore-gateway/internal/auth/handlers"
	"api--sigacore-gateway/internal/auth/services"
	db "api--sigacore-gateway/internal/db/sqlc"
	"api--sigacore-gateway/internal/shared/middleware"
	token2 "api--sigacore-gateway/internal/token"
	"api--sigacore-gateway/internal/util"
)

type AuthServer struct {
	config      util.Config
	authHandler *handlers.AuthHandler
	tokenMaker  token2.Maker
	router      *gin.Engine
}

func NewAuthServer(cfg util.Config, store db.Store) (*AuthServer, error) {
	ctx := context.Background()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		err := v.RegisterValidation("username", validUsername)
		if err != nil {
			return nil, fmt.Errorf("RegisterValidationCtx: %w", err)
		}
	}

	tokenMaker, err := token2.NewPasetoMaker(cfg.TokenSymmetricKey)
	conn, err := pgxpool.New(ctx, cfg.ConnStr)
	if err != nil {
		return nil, err
	}

	authService := services.NewAuthService(store, tokenMaker, cfg)
	authHandler := handlers.NewAuthHandler(authService, tokenMaker, conn, cfg)

	server := &AuthServer{
		config:      cfg,
		authHandler: authHandler,
		tokenMaker:  tokenMaker,
	}

	server.setupRoutes()
	return server, nil
}

func (s *AuthServer) setupRoutes() {
	router := gin.Default()

	// Rotas públicas
	router.POST("/users", s.authHandler.CreateUser)
	router.POST("/users/login", s.authHandler.LoginUser)
	router.POST("/token/renew", s.authHandler.RenewAccessToken)

	// Rotas protegidas
	authRoutes := router.Group("/").Use(middleware.AuthMiddleware(s.tokenMaker))
	authRoutes.GET("/users/:username", s.authHandler.GetUser)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Auth service is healthy"})
	})

	s.router = router
}

func (s *AuthServer) Start(address string) error {
	return s.router.Run(address)
}
