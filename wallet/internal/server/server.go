package server

import (
	"github.com/gin-gonic/gin"
	"wallet/internal/handlers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", handlers.HealthHandler)

	return r
}
