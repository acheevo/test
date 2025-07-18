package auth_integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/acheevo/test/internal/auth/domain"
	userDomain "github.com/acheevo/test/internal/user/domain"
	"github.com/acheevo/test/tests/integration/shared"
)

func TestAuthIntegration(t *testing.T) {
	deps := shared.SetupTestDependencies(t)
	defer deps.Cleanup(t)

	deps.SetupAuthRoutes()

	t.Run("User Registration", func(t *testing.T) {
		registerReq := domain.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		reqBody, _ := json.Marshal(registerReq)
		req := shared.MakeRequest(http.MethodPost, "/api/auth/register", reqBody)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var user userDomain.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, registerReq.Email, user.Email)
		assert.Equal(t, registerReq.Name, user.Name)
		assert.Equal(t, userDomain.RoleUser, user.Role)
		assert.Empty(t, user.Password) // Password should not be returned
	})

	t.Run("User Login", func(t *testing.T) {
		// First register a user
		registerReq := domain.RegisterRequest{
			Email:    "login@example.com",
			Password: "password123",
			Name:     "Login User",
		}

		reqBody, _ := json.Marshal(registerReq)
		req := shared.MakeRequest(http.MethodPost, "/api/auth/register", reqBody)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Now login
		loginReq := domain.LoginRequest{
			Email:    "login@example.com",
			Password: "password123",
		}

		reqBody, _ = json.Marshal(loginReq)
		req = shared.MakeRequest(http.MethodPost, "/api/auth/login", reqBody)

		w = httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var loginResp domain.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &loginResp)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResp.Token)
		assert.Equal(t, loginReq.Email, loginResp.User.Email)
		assert.Equal(t, userDomain.RoleUser, loginResp.User.Role)
	})

	t.Run("Invalid Login", func(t *testing.T) {
		loginReq := domain.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "wrongpassword",
		}

		reqBody, _ := json.Marshal(loginReq)
		req := shared.MakeRequest(http.MethodPost, "/api/auth/login", reqBody)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Protected Route Access", func(t *testing.T) {
		token := shared.CreateAndLoginUser(
			t, deps, "protected@example.com", "password123", "Protected User", userDomain.RoleUser,
		)

		// Access protected route
		req := shared.MakeAuthenticatedRequest(http.MethodGet, "/api/users/me", token, nil)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user userDomain.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, "protected@example.com", user.Email)
	})

	t.Run("Protected Route Without Token", func(t *testing.T) {
		req := shared.MakeRequest(http.MethodGet, "/api/users/me", nil)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
