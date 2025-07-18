package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/acheevo/test/internal/auth/repository"
	"github.com/acheevo/test/internal/auth/service"
	"github.com/acheevo/test/internal/auth/transport"
	"github.com/acheevo/test/internal/middleware"
	"github.com/acheevo/test/internal/shared/config"
	"github.com/acheevo/test/internal/shared/database"
	userRepository "github.com/acheevo/test/internal/user/repository"
	userService "github.com/acheevo/test/internal/user/service"
	userTransport "github.com/acheevo/test/internal/user/transport"
)

type Server struct {
	server *http.Server
	logger *zap.Logger
	db     *database.Database
}

func NewServer(logger *zap.Logger, cfg *config.Config) (*Server, error) {
	// Initialize database
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	userRepo := userRepository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize services
	authSvc := service.NewAuthService(userRepo, sessionRepo)
	userSvc := userService.NewUserService(userRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authSvc, logger)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		allowedHeaders := "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, " +
			"Authorization, accept, origin, Cache-Control, X-Requested-With"
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	setupRoutes(router, logger, userSvc, authSvc, authMiddleware)

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
	userSvc *userService.UserService,
	authSvc *service.AuthService,
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
		authHandler := transport.NewAuthHandler(authSvc, logger)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/logout", authHandler.Logout)
		}

		// User handlers
		userHandler := userTransport.NewUserHandler(userSvc, logger)

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
