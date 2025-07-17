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

func TestUserIntegration(t *testing.T) {
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
		}

		users := api.Group("/users")
		users.Use(authMiddleware.Authenticate)
		{
			users.GET("/me", userHandler.GetCurrentUser)
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUserByID)
			users.POST("", userHandler.CreateUser)
		}
	}

	// Helper function to create and login a user
	createAndLoginUser := func(email, password, name string, role models.UserRole) string {
		// Create user directly in database for admin user
		if role == models.RoleAdmin {
			_, err := userService.Create(email, password, name, role)
			require.NoError(t, err)
			
			// Login to get token
			loginReq := models.LoginRequest{
				Email:    email,
				Password: password,
			}

			reqBody, _ := json.Marshal(loginReq)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)

			var loginResp models.LoginResponse
			err = json.Unmarshal(w.Body.Bytes(), &loginResp)
			require.NoError(t, err)

			return loginResp.Token
		}

		// Register normal user
		registerReq := models.RegisterRequest{
			Email:    email,
			Password: password,
			Name:     name,
		}

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Login to get token
		loginReq := models.LoginRequest{
			Email:    email,
			Password: password,
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

		return loginResp.Token
	}

	t.Run("Get Current User", func(t *testing.T) {
		token := createAndLoginUser("current@example.com", "password123", "Current User", models.RoleUser)

		req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user models.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, "current@example.com", user.Email)
		assert.Equal(t, "Current User", user.Name)
		assert.Equal(t, models.RoleUser, user.Role)
	})

	t.Run("Get All Users - Admin Only", func(t *testing.T) {
		// Create admin user
		adminToken := createAndLoginUser("admin@example.com", "password123", "Admin User", models.RoleAdmin)

		// Create regular user
		createAndLoginUser("regular@example.com", "password123", "Regular User", models.RoleUser)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var users []models.User
		err := json.Unmarshal(w.Body.Bytes(), &users)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(users), 2) // At least admin and regular user
	})

	t.Run("Get All Users - Regular User Forbidden", func(t *testing.T) {
		userToken := createAndLoginUser("forbidden@example.com", "password123", "Forbidden User", models.RoleUser)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Create User - Admin Only", func(t *testing.T) {
		adminToken := createAndLoginUser("createadmin@example.com", "password123", "Create Admin", models.RoleAdmin)

		createUserReq := struct {
			Email    string           `json:"email"`
			Password string           `json:"password"`
			Name     string           `json:"name"`
			Role     models.UserRole  `json:"role"`
		}{
			Email:    "created@example.com",
			Password: "password123",
			Name:     "Created User",
			Role:     models.RoleUser,
		}

		reqBody, _ := json.Marshal(createUserReq)
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var user models.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, createUserReq.Email, user.Email)
		assert.Equal(t, createUserReq.Name, user.Name)
		assert.Equal(t, createUserReq.Role, user.Role)
	})

	t.Run("Create User - Regular User Forbidden", func(t *testing.T) {
		userToken := createAndLoginUser("createuser@example.com", "password123", "Create User", models.RoleUser)

		createUserReq := struct {
			Email    string           `json:"email"`
			Password string           `json:"password"`
			Name     string           `json:"name"`
			Role     models.UserRole  `json:"role"`
		}{
			Email:    "shouldnotcreate@example.com",
			Password: "password123",
			Name:     "Should Not Create",
			Role:     models.RoleUser,
		}

		reqBody, _ := json.Marshal(createUserReq)
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}