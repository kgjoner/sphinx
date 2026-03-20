package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/domains/identity/identrepo"
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

		var respData httpserver.Success[authcase.LoginOutput]
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

			var userInfo httpserver.Success[identity.UserView]
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

			resp2, err := ts.AuthenticatedRequest("POST", "/auth/refresh", nil, respData.Data.AccessToken)
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

	var loginData2 httpserver.Success[authcase.LoginOutput]
	json.NewDecoder(loginResp.Body).Decode(&loginData2)

	t.Run("should refresh token successfully", func(t *testing.T) {
		// Add a small delay to ensure different timestamps in JWT tokens
		time.Sleep(1 * time.Second)

		resp, err := ts.AuthenticatedRequest("POST", "/auth/refresh", nil, loginData2.Data.RefreshToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var refreshResp httpserver.Success[authcase.LoginOutput]
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

func TestIdentityProviderFlows(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()
	defer mocks.IdentityProviders.Close()

	presetProviderData := map[string]interface{}{
		"authorization": "Bearer " + mocks.Subject.Token,
	}

	t.Run("should reject authentication before creation", func(t *testing.T) {
		token := uuid.New().String()
		subID := uuid.New().String()
		subEmail := "ext" + GenerateEmail()
		mocks.IdentityProviders.AddSubjectTo(
			mocks.Provider.Name,
			token,
			subID,
			subEmail,
		)

		newSetProviderData := map[string]interface{}{
			"authorization": "Bearer " + token,
		}

		resp, err := ts.Request("POST", "/auth/external/"+mocks.Provider.Name, newSetProviderData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return unauthorized since there is no matching user yet
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		t.Run("should register user", func(t *testing.T) {
			resp, err := ts.Request("POST", "/user/external/"+mocks.Provider.Name, newSetProviderData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should successfully authenticate and create user
			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var respData httpserver.Success[identity.UserLeanView]
			err = json.NewDecoder(resp.Body).Decode(&respData)
			require.NoError(t, err)

			assert.NotEmpty(t, respData.Data.ID)

			// Verify user was created in database
			dao := identrepo.NewFactory().NewDAO(context.Background(), ts.server.GetBasePool().Connection())
			user, err := dao.GetUserByID(respData.Data.ID)
			require.NoError(t, err)
			require.NotNil(t, user)

			assert.Equal(t, subEmail, user.Email.String())

			t.Run("should login an already consented user without pass consent again", func(t *testing.T) {
				resp, err := ts.Request("POST", "/auth/external/"+mocks.Provider.Name, newSetProviderData, nil)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var loginResp httpserver.Success[authcase.LoginOutput]
				err = json.NewDecoder(resp.Body).Decode(&loginResp)
				require.NoError(t, err)

				assert.NotEmpty(t, loginResp.Data.AccessToken)
				assert.NotEmpty(t, loginResp.Data.RefreshToken)
				assert.Equal(t, respData.Data.ID.String(), loginResp.Data.UserID.String())
			})
		})
	})

	t.Run("should reject creating another user with existing external subject", func(t *testing.T) {
		// Ensure the external credential already exists. It may already fail due to previous test runs.
		// If it does, we can ignore the error. The real check is on the second request.
		preResp, err := ts.Request("POST", "/user/external/"+mocks.Provider.Name, presetProviderData, nil)
		require.NoError(t, err)
		preResp.Body.Close()

		resp, err := ts.Request("POST", "/user/external/"+mocks.Provider.Name, presetProviderData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return conflict due to existing external credential
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("should relate existing user", func(t *testing.T) {
		token := "valid-existing-456"
		subID := uuid.New().String()
		subEmail := "simple.external@sphinx.test"
		mocks.IdentityProviders.AddSubjectTo(
			mocks.Provider.Name,
			token,
			subID,
			subEmail,
		)

		// Login existing user
		resp1, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer resp1.Body.Close()

		var loginResp httpserver.Success[authcase.LoginOutput]
		err = json.NewDecoder(resp1.Body).Decode(&loginResp)
		require.NoError(t, err)
		require.NotEmpty(t, loginResp.Data.AccessToken)

		simpleProviderData := map[string]interface{}{
			"authorization": "Bearer " + token,
		}

		// Relate external credential to existing user
		resp, err := ts.Request("PUT", "/user/me/external/"+mocks.Provider.Name, simpleProviderData, map[string]string{
			"Authorization": "Bearer " + loginResp.Data.AccessToken,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should successfully relate existing user to external provider
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respData httpserver.Success[identity.ExternalCredentialView]
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, mocks.SimpleUser.ID.String(), respData.Data.UserID.String())
		assert.Equal(t, mocks.Provider.Name, respData.Data.ProviderName)
		assert.Equal(t, subID, respData.Data.ProviderSubjectID)
		assert.Equal(t, subEmail, respData.Data.ProviderAlias)

		t.Run("should login existing user with new external credential", func(t *testing.T) {
			resp, err := ts.Request("POST", "/auth/external/"+mocks.Provider.Name, simpleProviderData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var loginResp httpserver.Success[authcase.LoginOutput]
			err = json.NewDecoder(resp.Body).Decode(&loginResp)
			require.NoError(t, err)

			assert.NotEmpty(t, loginResp.Data.AccessToken)
			assert.NotEmpty(t, loginResp.Data.RefreshToken)
			assert.Equal(t, mocks.SimpleUser.ID, loginResp.Data.UserID)

			t.Run("should remove external credential", func(t *testing.T) {
				resp, err := ts.Request("DELETE", "/user/me/external/"+mocks.Provider.Name+"/"+subID, nil, map[string]string{
					"Authorization": "Bearer " + loginResp.Data.AccessToken,
				})
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			})
		})
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		presetProviderData := map[string]interface{}{
			"authorization": "Bearer invalid-token",
		}

		resp, err := ts.Request("POST", "/auth/external/"+mocks.Provider.Name, presetProviderData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return unauthorized for invalid token
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should handle provider error responses", func(t *testing.T) {
		resp, err := ts.Request("POST", "/auth/external/"+mocks.UnavailableProvider.Name, presetProviderData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return error status
		assert.True(t, resp.StatusCode >= 400)
	})
}
