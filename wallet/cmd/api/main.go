package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"wallet/internal/config"
	"wallet/internal/db"
	"wallet/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[WALLET] ERROR | %v\n", err)
	}

	dbPool, err := db.Connect(context.Background(), cfg.DSN)
	if err != nil {
		log.Fatalf("[WALLET] ERROR | %v\n", err)
	}
	defer dbPool.Close()

	if cfg.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	fmt.Fprintf(gin.DefaultWriter, "[WALLET] Listening on port :%s\n", cfg.Port)

	s := server.New(dbPool)

	if err := s.Run(cfg.Port); err != nil {
		log.Fatalf("[WALLET] ERROR | Server failed to start: %v\n", err)
	}
}
