package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"users/internal/handlers"
	"users/internal/middleware"
	"users/internal/repository"
)

const maxBodyBytes = 64 * 1024 // 64 KB

type Server struct {
	dbPool            *pgxpool.Pool
	router            *gin.Engine
	jwtSecret         string
	jwtInternalSecret string
}

func New(dbPool *pgxpool.Pool, jwtSecret, jwtInternalSecret string) *Server {
	s := &Server{
		dbPool:            dbPool,
		router:            gin.Default(),
		jwtSecret:         jwtSecret,
		jwtInternalSecret: jwtInternalSecret,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		c.Next()
	})

	userRepo := repository.NewUserRepository(s.dbPool)

	authHandler := handlers.NewAuthHandler(userRepo, s.jwtSecret)
	userHandler := handlers.NewUserHandler(userRepo)
	lookupHandler := handlers.NewLookupHandler(userRepo)

	s.router.GET("/health", handlers.HealthHandler)

	// Public auth routes
	s.router.POST("/users", authHandler.Register)
	s.router.POST("/sessions", authHandler.Login)

	// Authenticated user routes
	authed := s.router.Group("/users", middleware.JWT(s.jwtSecret))
	{
		authed.GET("/me", userHandler.GetMe)
		authed.PUT("/me", userHandler.UpdateMe)
	}

	// Internal route used by the wallet service (protected by internal secret)
	internal := s.router.Group("", middleware.JWT(s.jwtInternalSecret))
	{
		internal.GET("/users/:id", lookupHandler.GetByID)
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}
