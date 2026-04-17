package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"wallet/internal/config"
	"wallet/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	fmt.Fprintf(gin.DefaultWriter, "[WALLET] Listening on port :%s\n", cfg.Port)

	r := server.SetupRouter()

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("[WALLET] ERROR | Server failed to start: %v\n", err)
	}
}
