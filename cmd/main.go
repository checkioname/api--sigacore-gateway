package main

import (
	"context"
	"log"
	"sync"

	"api--sigacore-gateway/internal/auth"
	db "api--sigacore-gateway/internal/db/sqlc"
	"api--sigacore-gateway/internal/gateway"
	"api--sigacore-gateway/internal/util"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := pgxpool.New(ctx, config.ConnStr)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
		return
	}

	// Configurar banco de dados para o serviço de auth
	store := db.NewStore(conn)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}
	defer conn.Close()

	authServer, err := auth.NewAuthServer(config, store)
	if err != nil {
		log.Fatal("cannot create auth server:", err)
	}

	gatewayServer, err := gateway.NewGatewayServer(config)
	if err != nil {
		log.Fatal("cannot create gateway server:", err)
	}

	// Usar WaitGroup para rodar ambos os serviços
	var wg sync.WaitGroup
	wg.Add(2)

	// Iniciar Auth Service
	go func() {
		defer wg.Done()
		log.Printf("Starting Auth Service on %s", config.AuthServerAddress)
		if err := authServer.Start(config.AuthServerAddress); err != nil {
			log.Printf("Auth service error: %v", err)
		}
	}()

	// Iniciar Gateway
	go func() {
		defer wg.Done()
		log.Printf("Starting Gateway on %s", config.GatewayServerAddress)
		if err := gatewayServer.Start(); err != nil {
			log.Printf("Gateway error: %v", err)
		}
	}()

	log.Println("🚀 Sistema iniciado com sucesso!")
	log.Printf("📍 Auth Service: http://%s", config.AuthServerAddress)
	log.Printf("📍 Gateway: http://%s", config.GatewayServerAddress)

	// Aguardar ambos os serviços
	wg.Wait()
}
