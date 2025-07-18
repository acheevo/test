package shared

import (
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/acheevo/test/internal/auth/repository"
	"github.com/acheevo/test/internal/auth/service"
	"github.com/acheevo/test/internal/auth/transport"
	"github.com/acheevo/test/internal/middleware"
	"github.com/acheevo/test/internal/shared/testutil"
	userRepository "github.com/acheevo/test/internal/user/repository"
	userService "github.com/acheevo/test/internal/user/service"
	userTransport "github.com/acheevo/test/internal/user/transport"
)

// TestDependencies holds all the dependencies needed for integration tests
type TestDependencies struct {
	TestDB         *testutil.TestDB
	UserRepo       *userRepository.UserRepository
	SessionRepo    *repository.SessionRepository
	AuthService    *service.AuthService
	UserService    *userService.UserService
	AuthHandler    *transport.AuthHandler
	UserHandler    *userTransport.UserHandler
	AuthMiddleware *middleware.AuthMiddleware
	Router         *gin.Engine
	Logger         *zap.Logger
}

// SetupTestDependencies creates and configures all test dependencies
func SetupTestDependencies(t *testing.T) *TestDependencies {
	// Setup test database
	testDB := testutil.SetupTestDB(t)

	// Setup repositories
	userRepo := userRepository.NewUserRepository(testDB.Database)
	sessionRepo := repository.NewSessionRepository(testDB.Database)

	// Setup services
	authSvc := service.NewAuthService(userRepo, sessionRepo)
	userSvc := userService.NewUserService(userRepo)

	// Setup handlers
	logger := zap.NewNop()
	authHandler := transport.NewAuthHandler(authSvc, logger)
	userHandler := userTransport.NewUserHandler(userSvc, logger)
	authMiddleware := middleware.NewAuthMiddleware(authSvc, logger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	return &TestDependencies{
		TestDB:         testDB,
		UserRepo:       userRepo,
		SessionRepo:    sessionRepo,
		AuthService:    authSvc,
		UserService:    userSvc,
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		AuthMiddleware: authMiddleware,
		Router:         router,
		Logger:         logger,
	}
}

// Cleanup cleans up test resources
func (deps *TestDependencies) Cleanup(t *testing.T) {
	deps.TestDB.Cleanup(t)
}

// SetupAuthRoutes configures authentication routes for testing
func (deps *TestDependencies) SetupAuthRoutes() {
	api := deps.Router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", deps.AuthHandler.Register)
			auth.POST("/login", deps.AuthHandler.Login)
			auth.POST("/logout", deps.AuthHandler.Logout)
		}

		users := api.Group("/users")
		users.Use(deps.AuthMiddleware.Authenticate)
		{
			users.GET("/me", deps.UserHandler.GetCurrentUser)
		}
	}
}

// SetupUserRoutes configures user management routes for testing
func (deps *TestDependencies) SetupUserRoutes() {
	api := deps.Router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", deps.AuthHandler.Register)
			auth.POST("/login", deps.AuthHandler.Login)
		}

		users := api.Group("/users")
		users.Use(deps.AuthMiddleware.Authenticate)
		{
			users.GET("/me", deps.UserHandler.GetCurrentUser)
			users.GET("", deps.UserHandler.GetUsers)
			users.GET("/:id", deps.UserHandler.GetUserByID)
			users.POST("", deps.UserHandler.CreateUser)
		}
	}
}
