package main

import (
	"context"
	"log"
	"os"

	"api--sigacore-gateway/api"
	db "api--sigacore-gateway/db/sqlc"
	"api--sigacore-gateway/util"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.ConnStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer connPool.Close()

	store := db.NewStore(connPool)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	log.Printf("Starting Auth Service on %s", config.AuthServerAddress)
	err = server.Start(config.AuthServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
