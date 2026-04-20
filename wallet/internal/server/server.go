package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"wallet/internal/handlers"
	"wallet/internal/middleware"
	"wallet/internal/repository"
)

const maxBodyBytes = 64 * 1024 // 64 KB

type Server struct {
	dbPool    *pgxpool.Pool
	router    *gin.Engine
	jwtSecret string
}

func New(dbPool *pgxpool.Pool, jwtSecret string) *Server {
	s := &Server{
		dbPool:    dbPool,
		router:    gin.Default(),
		jwtSecret: jwtSecret,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		c.Next()
	})

	walletRepo := repository.NewWalletRepository(s.dbPool)
	transactionRepo := repository.NewTransactionRepository(s.dbPool)

	walletHandler := handlers.NewWalletHandler(walletRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionRepo)

	s.router.GET("/health", handlers.HealthHandler)

	wallets := s.router.Group("/wallets", middleware.JWT(s.jwtSecret))
	{
		wallets.GET("", walletHandler.List)
		wallets.GET("/:id", walletHandler.GetByID)
		wallets.POST("", walletHandler.Create)
		wallets.PUT("/:id", walletHandler.UpdateDescription)
		wallets.POST("/:id/transactions", transactionHandler.Create)
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}
