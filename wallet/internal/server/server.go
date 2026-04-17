package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"wallet/internal/handlers"
)

type Server struct {
	dbPool *pgxpool.Pool
	router *gin.Engine
}

func New(dbPool *pgxpool.Pool) *Server {
	s := &Server{
		dbPool: dbPool,
		router: gin.Default(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", handlers.HealthHandler)
}

func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}
