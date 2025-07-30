package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/config"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	oauthcase "github.com/kgjoner/sphinx/internal/domains/auth/cases/oauth/provider"
	"github.com/kgjoner/sphinx/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthAuthorizationFlow(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	// First, login to get an access token for the issueGrant endpoint
	loginData := factory.SimpleUserLoginData()
	loginResp, err := ts.Request("POST", "/auth/login", loginData, nil)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var loginResult presenter.Success[authcase.LoginOutput]
	err = json.NewDecoder(loginResp.Body).Decode(&loginResult)
	require.NoError(t, err)
	accessToken := loginResult.Data.AccessToken

	// Test 1: Issue Authorization Grant
	t.Run("should issue authorization grant successfully", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":      "authorization_code",
			"client_id":       mocks.CommonApplication.Id.String(),
			"redirect_uri":    mocks.CommonRedirectUri,
			"consent_granted": true,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var grantResp presenter.Success[oauthcase.IssueGrantOutput]
		err = json.NewDecoder(resp.Body).Decode(&grantResp)
		require.NoError(t, err)

		assert.NotEmpty(t, grantResp.Data.Code)
		assert.Equal(t, mocks.CommonRedirectUri, grantResp.Data.RedirectUri)
		assert.WithinDuration(t, time.Now().Add(time.Duration(config.Env.AUTH_GRANT_LIFETIME_IN_SEC)*time.Second), grantResp.Data.ExpiresAt, time.Minute)

		// Test 2: Exchange Authorization Grant for tokens
		t.Run("should exchange grant for tokens successfully", func(t *testing.T) {
			exchangeData := map[string]interface{}{
				"grant_type":    "authorization_code",
				"code":          grantResp.Data.Code,
				"redirect_uri":  mocks.CommonRedirectUri,
				"client_id":     mocks.CommonApplication.Id.String(),
				"client_secret": mocks.CommonAppSecret,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var tokenResp presenter.Success[authcase.LoginOutput]
			err = json.NewDecoder(resp.Body).Decode(&tokenResp)
			require.NoError(t, err)

			assert.NotEmpty(t, tokenResp.Data.AccessToken)
			assert.NotEmpty(t, tokenResp.Data.RefreshToken)
			assert.Equal(t, mocks.SimpleUserAccount.Id, tokenResp.Data.AccountId)
			assert.Equal(t, config.Env.JWT.ACCESS_LIFETIME_IN_SEC, tokenResp.Data.ExpiresIn)
		})

		// Test 3: Exchange with PKCE (code_verifier instead of client_secret)
		t.Run("should exchange grant with PKCE successfully", func(t *testing.T) {
			codeVerifier := "S256_code_challenge_example"
			hashedVerifier := sha256.Sum256([]byte(codeVerifier))
			codeChallenge := base64.RawURLEncoding.EncodeToString(hashedVerifier[:])
			// First issue a new grant
			grantData := map[string]interface{}{
				"grant_type":            "authorization_code",
				"client_id":             mocks.CommonApplication.Id.String(),
				"redirect_uri":          mocks.CommonRedirectUri,
				"code_challenge":        codeChallenge,
				"code_challenge_method": "S256",
			}

			grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
			require.NoError(t, err)
			defer grantResp.Body.Close()

			var newGrant presenter.Success[oauthcase.IssueGrantOutput]
			err = json.NewDecoder(grantResp.Body).Decode(&newGrant)
			require.NoError(t, err)

			// Exchange with code_verifier
			exchangeData := map[string]interface{}{
				"grant_type":    "authorization_code",
				"code":          newGrant.Data.Code,
				"redirect_uri":  mocks.CommonRedirectUri,
				"client_id":     mocks.CommonApplication.Id.String(),
				"code_verifier": codeVerifier,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var tokenResp presenter.Success[authcase.LoginOutput]
			err = json.NewDecoder(resp.Body).Decode(&tokenResp)
			require.NoError(t, err)

			assert.NotEmpty(t, tokenResp.Data.AccessToken)
			assert.NotEmpty(t, tokenResp.Data.RefreshToken)
		})
	})
}

func TestOAuthAuthorizationErrors(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	// Login to get access token for testing
	loginData := factory.SimpleUserLoginData()
	loginResp, err := ts.Request("POST", "/auth/login", loginData, nil)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var loginResult presenter.Success[authcase.LoginOutput]
	err = json.NewDecoder(loginResp.Body).Decode(&loginResult)
	require.NoError(t, err)
	accessToken := loginResult.Data.AccessToken

	// Unauthenticated Errors
	t.Run("should reject issueGrant without authentication", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":      "authorization_code",
			"client_id":       mocks.CommonApplication.Id.String(),
			"redirect_uri":    mocks.CommonRedirectUri,
			"consent_granted": true,
		}

		resp, err := ts.Request("POST", "/oauth/authorize", grantData, nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Forbidden Errors
	t.Run("should reject issueGrant with no consent provided", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":   "authorization_code",
			"client_id":    mocks.CommonApplication.Id.String(),
			"redirect_uri": mocks.CommonRedirectUri,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Unprocessable Entity Errors
	t.Run("should reject issueGrant with invalid grant_type", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":      "invalid_grant_type",
			"client_id":       mocks.CommonApplication.Id.String(),
			"redirect_uri":    mocks.CommonRedirectUri,
			"consent_granted": true,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject issueGrant with invalid redirect_uri", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":      "authorization_code",
			"client_id":       mocks.CommonApplication.Id.String(),
			"redirect_uri":    "invalid-uri",
			"consent_granted": true,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	// Bad Request Errors
	t.Run("should reject issueGrant with unknown client_id", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":      "authorization_code",
			"client_id":       uuid.New().String(),
			"redirect_uri":    mocks.CommonRedirectUri,
			"consent_granted": true,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should reject issueGrant with not registered redirect_uri", func(t *testing.T) {
		grantData := map[string]interface{}{
			"grant_type":      "authorization_code",
			"client_id":       mocks.CommonApplication.Id.String(),
			"redirect_uri":    "https://unregistered-uri.com/callback",
			"consent_granted": true,
		}

		resp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should reject invalid exchangeGrant requests", func(t *testing.T) {
		// Grant must be issued for each individual test because once it has been tried to exchange,
		// it cannot be reused
		
		// Unprocessable Entity Errors
		t.Run("should reject exchangeGrant with invalid grant_type", func(t *testing.T) {
			// First issue a new grant
			grantData := map[string]interface{}{
				"grant_type":      "authorization_code",
				"client_id":       mocks.CommonApplication.Id.String(),
				"redirect_uri":    mocks.CommonRedirectUri,
				"consent_granted": true,
			}

			grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
			require.NoError(t, err)
			defer grantResp.Body.Close()

			assert.Equal(t, http.StatusOK, grantResp.StatusCode)

			var grant presenter.Success[oauthcase.IssueGrantOutput]
			err = json.NewDecoder(grantResp.Body).Decode(&grant)
			require.NoError(t, err)

			// Try to exchange with invalid grant_type
			exchangeData := map[string]interface{}{
				"grant_type":    "invalid_grant_type",
				"code":          grant.Data.Code,
				"client_id":     mocks.CommonApplication.Id.String(),
				"client_secret": mocks.CommonAppSecret,
				"redirect_uri":  mocks.CommonRedirectUri,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		})

		t.Run("should reject exchangeGrant without client_secret or code_verifier", func(t *testing.T) {
			// First issue a new grant
			grantData := map[string]interface{}{
				"grant_type":      "authorization_code",
				"client_id":       mocks.CommonApplication.Id.String(),
				"redirect_uri":    mocks.CommonRedirectUri,
				"consent_granted": true,
			}

			grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
			require.NoError(t, err)
			defer grantResp.Body.Close()

			assert.Equal(t, http.StatusOK, grantResp.StatusCode)

			var grant presenter.Success[oauthcase.IssueGrantOutput]
			err = json.NewDecoder(grantResp.Body).Decode(&grant)
			require.NoError(t, err)

			// Try to exchange without client_secret or code_verifier
			exchangeData := map[string]interface{}{
				"grant_type":   "authorization_code",
				"code":         grant.Data.Code,
				"client_id":    mocks.CommonApplication.Id.String(),
				"redirect_uri": mocks.CommonRedirectUri,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		})

		// Unauthorized Errors
		t.Run("should reject exchangeGrant with invalid code", func(t *testing.T) {
			exchangeData := map[string]interface{}{
				"grant_type":    "authorization_code",
				"code":          "invalid_code",
				"client_id":     mocks.CommonApplication.Id.String(),
				"client_secret": mocks.CommonAppSecret,
				"redirect_uri":  mocks.CommonRedirectUri,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should reject exchangeGrant with wrong client_secret", func(t *testing.T) {
			// First issue a new grant
			grantData := map[string]interface{}{
				"grant_type":      "authorization_code",
				"client_id":       mocks.CommonApplication.Id.String(),
				"redirect_uri":    mocks.CommonRedirectUri,
				"consent_granted": true,
			}

			grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
			require.NoError(t, err)
			defer grantResp.Body.Close()

			assert.Equal(t, http.StatusOK, grantResp.StatusCode)

			var grant presenter.Success[oauthcase.IssueGrantOutput]
			err = json.NewDecoder(grantResp.Body).Decode(&grant)
			require.NoError(t, err)

			// Try to exchange with wrong client_secret	
			exchangeData := map[string]interface{}{
				"grant_type":    "authorization_code",
				"code":          grant.Data.Code,
				"client_id":     mocks.CommonApplication.Id.String(),
				"client_secret": "wrong_secret",
				"redirect_uri":  mocks.CommonRedirectUri,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should reject exchangeGrant with mismatched redirect_uri", func(t *testing.T) {
			// First issue a new grant
			grantData := map[string]interface{}{
				"grant_type":      "authorization_code",
				"client_id":       mocks.CommonApplication.Id.String(),
				"redirect_uri":    mocks.CommonRedirectUri,
				"consent_granted": true,
			}

			grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
			require.NoError(t, err)
			defer grantResp.Body.Close()

			assert.Equal(t, http.StatusOK, grantResp.StatusCode)

			var grant presenter.Success[oauthcase.IssueGrantOutput]
			err = json.NewDecoder(grantResp.Body).Decode(&grant)
			require.NoError(t, err)

			// Try to exchange with different redirect_uri
			exchangeData := map[string]interface{}{
				"grant_type":    "authorization_code",
				"code":          grant.Data.Code,
				"client_id":     mocks.CommonApplication.Id.String(),
				"client_secret": mocks.CommonAppSecret,
				"redirect_uri":  "https://different.com/callback",
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should reject exchangeGrant with mismatched application", func(t *testing.T) {
			// First issue a new grant
			grantData := map[string]interface{}{
				"grant_type":      "authorization_code",
				"client_id":       mocks.CommonApplication.Id.String(),
				"redirect_uri":    mocks.CommonRedirectUri,
				"consent_granted": true,
			}

			grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
			require.NoError(t, err)
			defer grantResp.Body.Close()

			assert.Equal(t, http.StatusOK, grantResp.StatusCode)

			var grant presenter.Success[oauthcase.IssueGrantOutput]
			err = json.NewDecoder(grantResp.Body).Decode(&grant)
			require.NoError(t, err)

			// Try to exchange with a different application credentials (id and secret)
			exchangeData := map[string]interface{}{
				"grant_type":    "authorization_code",
				"code":          grant.Data.Code,
				"client_id":     mocks.RootApplication.Id.String(),
				"client_secret": mocks.RootAppSecret,
				"redirect_uri":  mocks.CommonRedirectUri,
			}

			resp, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
}

func TestOAuthGrantExpiration(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	// Login to get access token
	loginData := factory.SimpleUserLoginData()
	loginResp, err := ts.Request("POST", "/auth/login", loginData, nil)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var loginResult presenter.Success[authcase.LoginOutput]
	err = json.NewDecoder(loginResp.Body).Decode(&loginResult)
	require.NoError(t, err)
	accessToken := loginResult.Data.AccessToken

	t.Run("should not allow grant reuse", func(t *testing.T) {
		// Issue grant
		grantData := map[string]interface{}{
			"grant_type":      "authorization_code",
			"client_id":       mocks.CommonApplication.Id.String(),
			"redirect_uri":    mocks.CommonRedirectUri,
			"consent_granted": true,
		}

		grantResp, err := ts.AuthenticatedRequest("POST", "/oauth/authorize", grantData, accessToken)
		require.NoError(t, err)
		defer grantResp.Body.Close()

		var grant presenter.Success[oauthcase.IssueGrantOutput]
		err = json.NewDecoder(grantResp.Body).Decode(&grant)
		require.NoError(t, err)

		// Exchange grant first time
		exchangeData := map[string]interface{}{
			"grant_type":    "authorization_code",
			"code":          grant.Data.Code,
			"redirect_uri":  mocks.CommonRedirectUri,
			"client_id":     mocks.CommonApplication.Id.String(),
			"client_secret": mocks.CommonAppSecret,
		}

		resp1, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
		require.NoError(t, err)
		defer resp1.Body.Close()

		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		// Try to exchange the same grant again
		resp2, err := ts.Request("POST", "/oauth/token", exchangeData, nil)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)
	})
}
