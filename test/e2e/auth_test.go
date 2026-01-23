package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/config"
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
		assert.NotEmpty(t, respData.Data.UserID)

		// Test 2: Get User Info with Token
		t.Run("should get user info with valid token", func(t *testing.T) {
			resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, respData.Data.AccessToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var userInfo presenter.Success[auth.UserPrivateView]
			err = json.NewDecoder(resp.Body).Decode(&userInfo)
			require.NoError(t, err)

			// Verify against seeded data
			seedData := ts.GetSeedData()
			assert.Equal(t, seedData.SimpleUser.Email.String(), userInfo.Data.Email.String())
			assert.Equal(t, seedData.SimpleUser.Username, userInfo.Data.Username)
			assert.Equal(t, seedData.SimpleUser.Phone.String(), userInfo.Data.Phone.String())
			assert.Equal(t, seedData.SimpleUser.Document.String(), userInfo.Data.Document.String())
		})

		// Test 3: Logout
		t.Run("should logout successfully", func(t *testing.T) {
			resp, err := ts.AuthenticatedRequest("POST", "/auth/logout", nil, respData.Data.AccessToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			resp2, err := ts.AuthenticatedRequest("GET", "/user/me", nil, respData.Data.AccessToken)
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
			"entry":    "nonexistent@sphinx.test",
			"password": "wrongpassword",
		}

		resp, err := ts.Request("POST", "/auth/login", loginData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should reject requests with invalid token", func(t *testing.T) {
		resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, "invalid-token")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestRefreshToken(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()
	userData := factory.RandomUser()

	// Create user
	resp1, err := ts.Request("POST", "/user", userData, nil)
	require.NoError(t, err)
	resp1.Body.Close()

	// Login to get tokens
	loginData := factory.CreateLoginData(userData["email"].(string), userData["password"].(string))
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
		resp1, err := ts.AuthenticatedRequest("GET", "/user/me", nil, token1)
		require.NoError(t, err)
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		resp2, err := ts.AuthenticatedRequest("GET", "/user/me", nil, token2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
	})
}

func TestExternalAuthenticationWithMockProvider(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	// Setup mock external auth provider
	mockAuthManager := mocks.NewMockExternalAuthManager()
	defer mockAuthManager.Close()

	// Create a test provider
	testEmail := "externalprovider@sphinx.test"
	testProvider := mockAuthManager.SetupTestProvider(
		"test-provider",
		"valid-token-123",
		"sub-123",
		testEmail,
	)

	// Setup the configuration with the mock provider
	originalProviders := config.Env.EXTERNAL_AUTH_PROVIDERS
	defer func() {
		config.Env.EXTERNAL_AUTH_PROVIDERS = originalProviders
	}()
	config.Env.EXTERNAL_AUTH_PROVIDERS = mockAuthManager.GetConfigs()

	t.Run("should reject creation without consent", func(t *testing.T) {
		email := factory.RandomUser()["email"].(string)
		token := "valid-creation-123"

		testProvider.AddValidToken(token, &mocks.MockAuthSubject{
			ID:    "creation-123",
			Email: email,
		})

		externalAuthData := map[string]interface{}{
			"providerName":        testProvider.Name,
			"authorizationHeader": "Bearer " + token,
		}

		resp, err := ts.Request("POST", "/auth/external", externalAuthData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return forbidden due to lack of consent
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("should create user with consent", func(t *testing.T) {
		externalAuthData := map[string]interface{}{
			"providerName":    testProvider.Name,
			"consentCreation": true,
		}

		resp, err := ts.Request("POST", "/auth/external", externalAuthData, map[string]string{
			"Authorization": "Bearer valid-token-123",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should successfully authenticate and create user
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respData presenter.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.NotEmpty(t, respData.Data.AccessToken)
		assert.NotEmpty(t, respData.Data.RefreshToken)
		assert.NotEmpty(t, respData.Data.UserID)

		// Verify user was created in database
		dao := ts.server.GetBasePool().NewDAO(context.Background())
		user, err := dao.GetUserByID(respData.Data.UserID)
		require.NoError(t, err)
		require.NotNil(t, user)

		assert.Equal(t, testEmail, user.Email.String())

		t.Run("should login an already consented user without pass consent again", func(t *testing.T) {
			externalAuthData := map[string]interface{}{
				"providerName": testProvider.Name,
			}

			resp, err := ts.Request("POST", "/auth/external", externalAuthData, map[string]string{
				"Authorization": "Bearer valid-token-123",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var loginResp presenter.Success[authcase.LoginOutput]
			err = json.NewDecoder(resp.Body).Decode(&loginResp)
			require.NoError(t, err)

			assert.NotEmpty(t, loginResp.Data.AccessToken)
			assert.NotEmpty(t, loginResp.Data.RefreshToken)
			assert.NotEmpty(t, loginResp.Data.UserID)
		})
	})

	t.Run("should reject relating existing user without consent", func(t *testing.T) {
		// First create a user
		userData := factory.RandomUser()
		email := userData["email"].(string)

		resp1, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		resp1.Body.Close()

		// Try to authenticate with external provider using same email but no consent
		token := "valid-existing-123"
		testProvider.AddValidToken(token, &mocks.MockAuthSubject{
			ID:    "existing-123",
			Email: email,
		})

		externalAuthData := map[string]interface{}{
			"providerName":        testProvider.Name,
			"authorizationHeader": "Bearer " + token,
		}

		resp, err := ts.Request("POST", "/auth/external", externalAuthData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return forbidden due to lack of consent
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("should relate existing user with consent", func(t *testing.T) {
		// First create a user
		userData := factory.RandomUser()
		email := userData["email"].(string)

		resp1, err := ts.Request("POST", "/user", userData, nil)
		require.NoError(t, err)
		resp1.Body.Close()

		// Try to authenticate with external provider using same email
		token := "valid-existing-456"
		testProvider.AddValidToken(token, &mocks.MockAuthSubject{
			ID:    "sub-existing-123",
			Email: email,
		})
		externalAuthData := map[string]interface{}{
			"providerName":        testProvider.Name,
			"consentRelation":     true,
			"authorizationHeader": "Bearer " + token,
		}

		resp, err := ts.Request("POST", "/auth/external", externalAuthData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should successfully relate existing user to external provider
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respData presenter.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.NotEmpty(t, respData.Data.AccessToken)
		assert.NotEmpty(t, respData.Data.RefreshToken)
		assert.NotEmpty(t, respData.Data.UserID)

		t.Run("should login a related existing user without pass consent again", func(t *testing.T) {
			externalAuthData := map[string]interface{}{
				"providerName": testProvider.Name,
			}

			resp, err := ts.Request("POST", "/auth/external", externalAuthData, map[string]string{
				"Authorization": "Bearer " + token,
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var loginResp presenter.Success[authcase.LoginOutput]
			err = json.NewDecoder(resp.Body).Decode(&loginResp)
			require.NoError(t, err)

			assert.NotEmpty(t, loginResp.Data.AccessToken)
			assert.NotEmpty(t, loginResp.Data.RefreshToken)
			assert.NotEmpty(t, loginResp.Data.UserID)
		})
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		externalAuthData := map[string]interface{}{
			"providerName":        "test-provider",
			"consentCreation":     true,
			"email":               "externalprovider2@sphinx.test",
			"authorizationHeader": "Bearer invalid-token",
		}

		resp, err := ts.Request("POST", "/auth/external", externalAuthData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return unauthorized for invalid token
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should handle provider error responses", func(t *testing.T) {
		// Configure the provider to return an error
		testProvider.SetError(http.StatusInternalServerError, "Provider temporarily unavailable")

		externalAuthData := map[string]interface{}{
			"providerName":        "test-provider",
			"consentCreation":     true,
			"email":               "externalprovider2@sphinx.test",
			"authorizationHeader": "Bearer valid-token-123",
		}

		resp, err := ts.Request("POST", "/auth/external", externalAuthData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return error status
		assert.True(t, resp.StatusCode >= 400)

		// Reset the provider to normal operation
		testProvider.ClearError()
	})
}
