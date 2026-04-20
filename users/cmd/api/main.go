package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"users/internal/config"
	"users/internal/db"
	"users/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[USERS] ERROR | %v\n", err)
	}

	dbPool, err := db.Connect(context.Background(), cfg.DSN)
	if err != nil {
		log.Fatalf("[USERS] ERROR | %v\n", err)
	}
	defer dbPool.Close()

	err = db.Migrate(cfg.DSN)
	if err != nil {
		log.Fatalf("[USERS] ERROR | migrations: %v\n", err)
	}

	if cfg.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	fmt.Fprintf(gin.DefaultWriter, "[USERS] Listening on port :%s\n", cfg.Port)

	s := server.New(dbPool, cfg.JWTSecret, cfg.JWTInternalSecret)

	err = s.Run(cfg.Port)
	if err != nil {
		log.Fatalf("[USERS] ERROR | Server failed to start: %v\n", err)
	}
}
