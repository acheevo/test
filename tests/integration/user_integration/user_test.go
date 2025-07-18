package user_integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/acheevo/test/internal/user/domain"
	"github.com/acheevo/test/tests/integration/shared"
)

func TestUserIntegration(t *testing.T) {
	deps := shared.SetupTestDependencies(t)
	defer deps.Cleanup(t)

	deps.SetupUserRoutes()

	t.Run("Get Current User", func(t *testing.T) {
		token := shared.CreateAndLoginUser(
			t, deps, "current@example.com", "password123", "Current User", userDomain.RoleUser,
		)

		req := shared.MakeAuthenticatedRequest(http.MethodGet, "/api/users/me", token, nil)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user userDomain.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, "current@example.com", user.Email)
		assert.Equal(t, "Current User", user.Name)
		assert.Equal(t, userDomain.RoleUser, user.Role)
	})

	t.Run("Get All Users - Admin Only", func(t *testing.T) {
		// Create admin user
		adminToken := shared.CreateAndLoginUser(
			t, deps, "admin@example.com", "password123", "Admin User", userDomain.RoleAdmin,
		)

		// Create regular user
		shared.CreateAndLoginUser(
			t, deps, "regular@example.com", "password123", "Regular User", userDomain.RoleUser,
		)

		req := shared.MakeAuthenticatedRequest(http.MethodGet, "/api/users", adminToken, nil)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var users []userDomain.User
		err := json.Unmarshal(w.Body.Bytes(), &users)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(users), 2) // At least admin and regular user
	})

	t.Run("Get All Users - Regular User Forbidden", func(t *testing.T) {
		userToken := shared.CreateAndLoginUser(
			t, deps, "forbidden@example.com", "password123", "Forbidden User", userDomain.RoleUser,
		)

		req := shared.MakeAuthenticatedRequest(http.MethodGet, "/api/users", userToken, nil)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Create User - Admin Only", func(t *testing.T) {
		adminToken := shared.CreateAndLoginUser(
			t, deps, "createadmin@example.com", "password123", "Create Admin", userDomain.RoleAdmin,
		)

		createUserReq := userDomain.CreateUserRequest{
			Email:    "created@example.com",
			Password: "password123",
			Name:     "Created User",
			Role:     userDomain.RoleUser,
		}

		reqBody, _ := json.Marshal(createUserReq)
		req := shared.MakeAuthenticatedRequest(http.MethodPost, "/api/users", adminToken, reqBody)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var user userDomain.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err)

		assert.Equal(t, createUserReq.Email, user.Email)
		assert.Equal(t, createUserReq.Name, user.Name)
		assert.Equal(t, createUserReq.Role, user.Role)
	})

	t.Run("Create User - Regular User Forbidden", func(t *testing.T) {
		userToken := shared.CreateAndLoginUser(
			t, deps, "createuser@example.com", "password123", "Create User", userDomain.RoleUser,
		)

		createUserReq := userDomain.CreateUserRequest{
			Email:    "shouldnotcreate@example.com",
			Password: "password123",
			Name:     "Should Not Create",
			Role:     userDomain.RoleUser,
		}

		reqBody, _ := json.Marshal(createUserReq)
		req := shared.MakeAuthenticatedRequest(http.MethodPost, "/api/users", userToken, reqBody)

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
