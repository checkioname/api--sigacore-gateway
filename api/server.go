package api

import (
	db "api--sigacore-gateway/db/sqlc"
	"api--sigacore-gateway/token"
	"api--sigacore-gateway/util"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/go-playground/validator/v10"
)

var ErrCouldNotParse = errors.New("could not parse body")

type Server struct {
	config util.Config
	store  db.Store
	token  token.Maker
	router *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("NewServer: %w", err)
	}

	server := &Server{config: config,
		store: store,
		token: tokenMaker,
	}

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		//exemplo de validacao (vai ser outra em producao
		err = v.RegisterValidation("currency", validCurrency)
		if err != nil {
			return nil, fmt.Errorf("RegisterValidationCtx: %w", err)
		}
	}

	server.router = server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() *gin.Engine {
	router := gin.Default()

	// Requests to the routes bellow will pass on this middleware before
	router.Group("/").Use(authMiddleware(s.token))

	return router
}

func (s *Server) Start(address string) error {
	err := s.router.Run(address)
	if err != nil {
		return fmt.Errorf("ServerStart: %w", err)
	}
	return nil
}
