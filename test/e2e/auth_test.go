package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	"github.com/kgjoner/sphinx/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullAuthenticationFlow(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.server.Close()

	factory := NewTestDataFactory()

	// Test 1: Login
	t.Run("should login with valid credentials", func(t *testing.T) {
		resp, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respData presenter.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.NotEmpty(t, respData.Data.AccessToken)
		assert.NotEmpty(t, respData.Data.RefreshToken)
		assert.NotEmpty(t, respData.Data.AccountId)

		// Test 2: Get Account Info with Token
		t.Run("should get account info with valid token", func(t *testing.T) {
			resp, err := ts.AuthenticatedRequest("GET", "/account", nil, respData.Data.AccessToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var accountInfo presenter.Success[auth.AccountPrivateView]
			err = json.NewDecoder(resp.Body).Decode(&accountInfo)
			require.NoError(t, err)

			assert.Equal(t, mocks.SimpleUserAccount.Email.String(), accountInfo.Data.Email.String())
			assert.Equal(t, mocks.SimpleUserAccount.Username, accountInfo.Data.Username)
			assert.Equal(t, mocks.SimpleUserAccount.Phone.String(), accountInfo.Data.Phone.String())
			assert.Equal(t, mocks.SimpleUserAccount.Document.String(), accountInfo.Data.Document.String())
		})

		// Test 3: Logout
		t.Run("should logout successfully", func(t *testing.T) {
			resp, err := ts.AuthenticatedRequest("POST", "/auth/logout", nil, respData.Data.AccessToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			resp2, err := ts.AuthenticatedRequest("GET", "/account", nil, respData.Data.AccessToken)
			require.NoError(t, err)
			defer resp2.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)
		})
	})
}

func TestAuthenticationErrors(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	t.Run("should reject invalid credentials", func(t *testing.T) {
		loginData := map[string]interface{}{
			"entry":    "nonexistent@example.com",
			"password": "wrongpassword",
		}

		resp, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should reject requests without app header", func(t *testing.T) {
		loginData := map[string]interface{}{
			"entry":    "test@example.com",
			"password": "password",
		}

		resp, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should reject requests with invalid token", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("GET", "/account", nil, "invalid-token")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestRefreshToken(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()
	accountData := factory.RandomAccount()

	// Create account
	resp1, err := ts.Request("POST", "/account", accountData, nil)
	require.NoError(t, err)
	resp1.Body.Close()

	// Login to get tokens
	loginData := factory.CreateLoginData(accountData["email"].(string), accountData["password"].(string))
	loginResp, err := ts.Request("POST", "/auth/login", loginData, nil)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var loginData2 presenter.Success[authcase.LoginOutput]
	json.NewDecoder(loginResp.Body).Decode(&loginData2)

	t.Run("should refresh token successfully", func(t *testing.T) {
		// Add a small delay to ensure different timestamps in JWT tokens
		time.Sleep(1 * time.Second)

		resp, err := ts.AuthenticatedRequest("POST", "/auth/refresh", nil, loginData2.Data.RefreshToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var refreshResp presenter.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp.Body).Decode(&refreshResp)
		require.NoError(t, err)

		assert.NotEmpty(t, refreshResp.Data.AccessToken)
		assert.NotEqual(t, loginData2.Data.AccessToken, refreshResp.Data.AccessToken)
		assert.NotEqual(t, loginData2.Data.RefreshToken, refreshResp.Data.RefreshToken)
	})
}

func TestSessionManagement(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	t.Run("should handle multiple sessions", func(t *testing.T) {
		loginData := factory.SimpleUserLoginData()
		// Login first session
		session1, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer session1.Body.Close()
		assert.Equal(t, http.StatusOK, session1.StatusCode)

		// Login second session
		session2, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer session2.Body.Close()
		assert.Equal(t, http.StatusOK, session2.StatusCode)

		// Both sessions should be valid
		var session1Data, session2Data map[string]interface{}
		json.NewDecoder(session1.Body).Decode(&session1Data)
		json.NewDecoder(session2.Body).Decode(&session2Data)

		token1 := session1Data["data"].(map[string]interface{})["accessToken"].(string)
		token2 := session2Data["data"].(map[string]interface{})["accessToken"].(string)

		assert.NotEqual(t, token1, token2)

		// Both tokens should work
		resp1, err := ts.AuthenticatedRequest("GET", "/account", nil, token1)
		require.NoError(t, err)
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		resp2, err := ts.AuthenticatedRequest("GET", "/account", nil, token2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
	})
}
