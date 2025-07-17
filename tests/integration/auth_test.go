package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/acheevo/test/internal/handlers"
	"github.com/acheevo/test/internal/middleware"
	"github.com/acheevo/test/internal/models"
	"github.com/acheevo/test/internal/repository"
	"github.com/acheevo/test/internal/services"
	"github.com/acheevo/test/internal/testutil"
)

func TestAuthIntegration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup(t)

	// Setup services
	userRepo := repository.NewUserRepository(testDB.Database)
	sessionRepo := repository.NewSessionRepository(testDB.Database)
	authService := services.NewAuthService(userRepo, sessionRepo)
	userService := services.NewUserService(userRepo)

	// Setup handlers
	logger := zap.NewNop()
	authHandler := handlers.NewAuthHandler(authService, logger)
	userHandler := handlers.NewUserHandler(userService, logger)
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		users := api.Group("/users")
		users.Use(authMiddleware.Authenticate)
		{
			users.GET("/me", userHandler.GetCurrentUser)
		}
	}

	t.Run("User Registration", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var user models.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, registerReq.Email, user.Email)
		assert.Equal(t, registerReq.Name, user.Name)
		assert.Equal(t, models.RoleUser, user.Role)
		assert.Empty(t, user.Password) // Password should not be returned
	})

	t.Run("User Login", func(t *testing.T) {
		// First register a user
		registerReq := models.RegisterRequest{
			Email:    "login@example.com",
			Password: "password123",
			Name:     "Login User",
		}

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Now login
		loginReq := models.LoginRequest{
			Email:    "login@example.com",
			Password: "password123",
		}

		reqBody, _ = json.Marshal(loginReq)
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var loginResp models.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &loginResp)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResp.Token)
		assert.Equal(t, loginReq.Email, loginResp.User.Email)
		assert.Equal(t, models.RoleUser, loginResp.User.Role)
	})

	t.Run("Invalid Login", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "wrongpassword",
		}

		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Protected Route Access", func(t *testing.T) {
		// Register and login to get token
		registerReq := models.RegisterRequest{
			Email:    "protected@example.com",
			Password: "password123",
			Name:     "Protected User",
		}

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		loginReq := models.LoginRequest{
			Email:    "protected@example.com",
			Password: "password123",
		}

		reqBody, _ = json.Marshal(loginReq)
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp models.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &loginResp)
		require.NoError(t, err)

		// Access protected route
		req = httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.Token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user models.User
		err = json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, loginReq.Email, user.Email)
	})

	t.Run("Protected Route Without Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}