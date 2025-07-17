package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/acheevo/test/internal/config"
	"github.com/acheevo/test/internal/handlers"
	"github.com/acheevo/test/internal/middleware"
	"github.com/acheevo/test/internal/repository"
	"github.com/acheevo/test/internal/services"
)

type Server struct {
	server *http.Server
	logger *zap.Logger
	db     *repository.Database
}

func NewServer(logger *zap.Logger, cfg *config.Config) (*Server, error) {
	// Initialize database
	db, err := repository.NewDatabase(cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo)
	userService := services.NewUserService(userRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	setupRoutes(router, logger, userService, authService, authMiddleware)

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server: server,
		logger: logger,
		db:     db,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Close database connection
	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database connection", zap.Error(err))
	}

	return s.server.Shutdown(ctx)
}

func setupRoutes(
	router *gin.Engine,
	logger *zap.Logger,
	userService *services.UserService,
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "test-api"})
	})

	// API routes group
	api := router.Group("/api")
	{
		// Auth handlers
		authHandler := handlers.NewAuthHandler(authService, logger)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/logout", authHandler.Logout)
		}

		// User handlers
		userHandler := handlers.NewUserHandler(userService, logger)

		// Protected routes with authentication middleware
		protected := api.Group("/users")
		protected.Use(authMiddleware.Authenticate)
		{
			protected.GET("/me", userHandler.GetCurrentUser)
			protected.GET("", userHandler.GetUsers)
			protected.GET("/:id", userHandler.GetUserByID)
			protected.POST("", userHandler.CreateUser)
		}
	}
}