package shared

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/acheevo/test/internal/auth/domain"
	userDomain "github.com/acheevo/test/internal/user/domain"
)

// CreateAndLoginUser is a helper function to create and login a user for testing
func CreateAndLoginUser(
	t *testing.T, deps *TestDependencies, email, password, name string, role userDomain.UserRole,
) string {
	// Create user directly in database for admin user
	if role == userDomain.RoleAdmin {
		_, err := deps.UserService.Create(email, password, name, role)
		require.NoError(t, err)

		// Login to get token
		loginReq := domain.LoginRequest{
			Email:    email,
			Password: password,
		}

		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		deps.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp domain.LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &loginResp)
		require.NoError(t, err)

		return loginResp.Token
	}

	// Register normal user
	registerReq := domain.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	reqBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Login to get token
	loginReq := domain.LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	deps.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp domain.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	return loginResp.Token
}

// MakeAuthenticatedRequest is a helper to make HTTP requests with authentication
func MakeAuthenticatedRequest(method, url, token string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

// MakeRequest is a helper to make HTTP requests without authentication
func MakeRequest(method, url string, body []byte) *http.Request {
	return MakeAuthenticatedRequest(method, url, "", body)
}
