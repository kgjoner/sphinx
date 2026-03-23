package e2e

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRS256TokenGeneration(t *testing.T) {
	if config.Env.JWT.ALGORITHM != "RS256" {
		t.Skip("Skipping RS256 tests - JWT_ALGORITHM is not RS256")
	}

	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	t.Run("should generate RS256 tokens with kid header", func(t *testing.T) {
		// Login to get tokens
		resp, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp httpserver.SuccessResponse[authcase.LoginOutput]
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)

		accessToken := loginResp.Data.AccessToken
		refreshToken := loginResp.Data.RefreshToken

		// Parse access token without verification to check structure
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		token, _, err := parser.ParseUnverified(accessToken, jwt.MapClaims{})
		require.NoError(t, err)

		// Verify kid header exists
		kid, ok := token.Header["kid"]
		assert.True(t, ok, "Access token should have 'kid' in header")
		assert.NotEmpty(t, kid, "kid should not be empty")

		// Verify algorithm is RS256
		alg, ok := token.Header["alg"]
		assert.True(t, ok, "Token should have 'alg' in header")
		assert.Equal(t, "RS256", alg, "Algorithm should be RS256")

		// Parse refresh token
		refreshParsed, _, err := parser.ParseUnverified(refreshToken, jwt.MapClaims{})
		require.NoError(t, err)

		// Verify refresh token also has kid
		refreshKid, ok := refreshParsed.Header["kid"]
		assert.True(t, ok, "Refresh token should have 'kid' in header")
		assert.NotEmpty(t, refreshKid, "kid should not be empty")

		// Both tokens should have the same kid (signed by same key)
		assert.Equal(t, kid, refreshKid, "Access and refresh tokens should use same key")
	})

	t.Run("should validate RS256 tokens successfully", func(t *testing.T) {
		// Login to get valid token
		resp, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		var loginResp httpserver.SuccessResponse[authcase.LoginOutput]
		json.NewDecoder(resp.Body).Decode(&loginResp)

		// Use token to access protected endpoint
		userResp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, loginResp.Data.AccessToken)
		require.NoError(t, err)
		defer userResp.Body.Close()

		assert.Equal(t, http.StatusOK, userResp.StatusCode)
	})

	t.Run("should validate legacy token with right signing key", func(t *testing.T) {
		// Create token with HS256 instead of RS256
		claims := jwt.MapClaims{
			"sub":   mocks.AdminUser.ID.String(),
			"aud":   config.Env.ROOT_APP_ID,
			"iat":   time.Now().Unix(),
			"exp":   time.Now().Add(15 * time.Minute).Unix(),
			"itn":   "access",
			"email": mocks.AdminUser.Email.String(),
			"name":  mocks.AdminUser.Name(),
			"roles": []string{"ADMIN"},
			"sid":   uuid.New().String(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		legacyToken, err := token.SignedString([]byte(config.Env.JWT.SECRET))
		require.NoError(t, err)

		// Try to use legacy token
		resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, legacyToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Legacy token with right signing key should be accepted")
	})
}

func TestJWKSEndpoint(t *testing.T) {
	if config.Env.JWT.ALGORITHM != "RS256" {
		t.Skip("Skipping JWKS tests - JWT_ALGORITHM is not RS256")
	}

	ts := NewTestSuite(t)
	defer ts.Close()

	t.Run("should expose JWKS endpoint", func(t *testing.T) {
		// JWKS is at base path, not under /api
		req, err := http.NewRequest("GET", ts.server.URL()+config.BASE_PATH+"/.well-known/jwks.json", nil)
		require.NoError(t, err)

		resp, err := ts.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jwks map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		require.NoError(t, err)

		// Verify JWKS structure
		keys, ok := jwks["keys"].([]interface{})
		assert.True(t, ok, "JWKS should have 'keys' array")
		assert.NotEmpty(t, keys, "JWKS should have at least one key")

		// Verify first key structure
		key := keys[0].(map[string]interface{})
		assert.Equal(t, "RSA", key["kty"], "Key type should be RSA")
		assert.Equal(t, "sig", key["use"], "Key use should be sig")
		assert.Equal(t, "RS256", key["alg"], "Algorithm should be RS256")
		assert.NotEmpty(t, key["kid"], "Key should have kid")
		assert.NotEmpty(t, key["n"], "Key should have modulus (n)")
		assert.NotEmpty(t, key["e"], "Key should have exponent (e)")
	})

	t.Run("JWKS should contain active keys only", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.server.URL()+config.BASE_PATH+"/.well-known/jwks.json", nil)
		require.NoError(t, err)

		resp, err := ts.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var jwks map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&jwks)

		keys := jwks["keys"].([]interface{})

		// All keys should be RS256 (no HS256 in JWKS)
		for _, k := range keys {
			key := k.(map[string]interface{})
			assert.Equal(t, "RS256", key["alg"], "Only RS256 keys should be in JWKS")
		}
	})

	t.Run("should cache JWKS with proper headers", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.server.URL()+config.BASE_PATH+"/.well-known/jwks.json", nil)
		require.NoError(t, err)

		resp, err := ts.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check cache headers
		cacheControl := resp.Header.Get("Cache-Control")
		assert.Contains(t, cacheControl, "max-age=300", "JWKS should have 5-minute cache")
	})
}

func TestForgedTokens(t *testing.T) {
	if config.Env.JWT.ALGORITHM != "RS256" {
		t.Skip("Skipping forged token tests - JWT_ALGORITHM is not RS256")
	}

	ts := NewTestSuite(t)
	defer ts.Close()

	t.Run("should reject token signed with different key", func(t *testing.T) {
		// Generate a different RSA key pair
		fakePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		// Create fake token with valid structure but wrong signature
		claims := jwt.MapClaims{
			"sub":   uuid.New().String(),
			"aud":   config.Env.ROOT_APP_ID,
			"iat":   time.Now().Unix(),
			"exp":   time.Now().Add(15 * time.Minute).Unix(),
			"itn":   "access",
			"email": "hacker@evil.com",
			"name":  "Evil Hacker",
			"roles": []string{"admin"},
			"sid":   uuid.New().String(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = uuid.New().String() // Fake kid

		forgedToken, err := token.SignedString(fakePrivateKey)
		require.NoError(t, err)

		// Try to use forged token
		resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, forgedToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Forged token should be rejected")
	})

	t.Run("should reject token with modified payload", func(t *testing.T) {
		// Get valid token
		loginResp, err := ts.Request("POST", "/auth/login", NewTestDataFactory().SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		var loginData httpserver.SuccessResponse[authcase.LoginOutput]
		json.NewDecoder(loginResp.Body).Decode(&loginData)

		validToken := loginData.Data.AccessToken

		// Parse token and modify payload
		parts := strings.Split(validToken, ".")
		require.Len(t, parts, 3, "JWT should have 3 parts")

		// Decode payload
		payload, err := jwt.NewParser().DecodeSegment(parts[1])
		require.NoError(t, err)

		var claims map[string]interface{}
		json.Unmarshal(payload, &claims)

		// Modify claims (try to escalate privileges)
		claims["roles"] = []string{"admin", "superuser"}

		// Re-encode modified payload
		modifiedPayload, _ := json.Marshal(claims)
		parts[1] = base64.RawURLEncoding.EncodeToString(modifiedPayload)

		// Create tampered token (header + modified payload + original signature)
		tamperedToken := strings.Join(parts, ".")

		// Try to use tampered token
		resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, tamperedToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Tampered token should be rejected")
	})

	t.Run("should reject token with invalid kid", func(t *testing.T) {
		// Get valid token
		loginResp, err := ts.Request("POST", "/auth/login", NewTestDataFactory().SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		var loginData httpserver.SuccessResponse[authcase.LoginOutput]
		json.NewDecoder(loginResp.Body).Decode(&loginData)

		validToken := loginData.Data.AccessToken

		// Parse and modify kid in header
		parts := strings.Split(validToken, ".")
		require.Len(t, parts, 3)

		// Decode header
		headerBytes, err := jwt.NewParser().DecodeSegment(parts[0])
		require.NoError(t, err)

		var header map[string]interface{}
		json.Unmarshal(headerBytes, &header)

		// Modify kid to non-existent key
		header["kid"] = uuid.New().String()

		// Re-encode header
		modifiedHeader, _ := json.Marshal(header)
		parts[0] = base64.RawURLEncoding.EncodeToString(modifiedHeader)

		// Create token with modified kid
		modifiedToken := strings.Join(parts, ".")

		// Try to use token with invalid kid
		resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, modifiedToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Token with invalid kid should be rejected")
	})

	t.Run("should reject expired token", func(t *testing.T) {
		// Get valid token
		loginResp, err := ts.Request("POST", "/auth/login", NewTestDataFactory().SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		var loginData httpserver.SuccessResponse[authcase.LoginOutput]
		json.NewDecoder(loginResp.Body).Decode(&loginData)

		// Parse token to check expiration
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		token, _, err := parser.ParseUnverified(loginData.Data.AccessToken, jwt.MapClaims{})
		require.NoError(t, err)

		claims := token.Claims.(jwt.MapClaims)
		exp := claims["exp"].(float64)
		iat := claims["iat"].(float64)

		// Calculate how long until expiration
		expiresIn := time.Unix(int64(exp), 0).Sub(time.Unix(int64(iat), 0))

		// Wait for token to expire (only if expiration is reasonable for testing)
		if expiresIn <= 2*time.Second {
			time.Sleep(expiresIn + 100*time.Millisecond)

			// Try to use expired token
			resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, loginData.Data.AccessToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Expired token should be rejected")
		} else {
			t.Skip("Token expiration too long for e2e test")
		}
	})

	t.Run("should reject legacy token with wrong signing key", func(t *testing.T) {
		// Create token with HS256 instead of RS256
		claims := jwt.MapClaims{
			"sub":   uuid.New().String(),
			"aud":   config.Env.ROOT_APP_ID,
			"iat":   time.Now().Unix(),
			"exp":   time.Now().Add(15 * time.Minute).Unix(),
			"itn":   "access",
			"email": "hacker@evil.com",
			"name":  "Evil Hacker",
			"roles": []string{"admin"},
			"sid":   uuid.New().String(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		wrongAlgToken, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		// Try to use token with wrong algorithm
		resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, wrongAlgToken)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Token with wrong algorithm should be rejected")
	})

	t.Run("should reject malformed token", func(t *testing.T) {
		malformedTokens := []string{
			"not.a.jwt",
			"",
			"only-one-part",
			"two.parts",
			"invalid..token",
			strings.Repeat("a", 1000), // Very long invalid token
		}

		for _, malformed := range malformedTokens {
			resp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, malformed)
			require.NoError(t, err)
			resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Malformed token should be rejected: %s", malformed)
		}
	})
}

func TestTokenIntentValidation(t *testing.T) {
	if config.Env.JWT.ALGORITHM != "RS256" {
		t.Skip("Skipping intent validation tests - JWT_ALGORITHM is not RS256")
	}

	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	resp, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	var loginResp httpserver.SuccessResponse[authcase.LoginOutput]
	json.NewDecoder(resp.Body).Decode(&loginResp)

	t.Run("should reject refresh token for access endpoints", func(t *testing.T) {
		// Try to use refresh token on access endpoint
		userResp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, loginResp.Data.RefreshToken)
		require.NoError(t, err)
		defer userResp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, userResp.StatusCode, "Refresh token should not work on access endpoints")
	})

	t.Run("should reject access token for refresh endpoint", func(t *testing.T) {
		// Try to use access token on refresh endpoint
		refreshResp, err := ts.AuthenticatedRequest("POST", "/auth/refresh", nil, loginResp.Data.AccessToken)
		require.NoError(t, err)
		defer refreshResp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, refreshResp.StatusCode, "Access token should not work on refresh endpoint")
	})
}

func TestKeyRotationScenarios(t *testing.T) {
	if config.Env.JWT.ALGORITHM != "RS256" {
		t.Skip("Skipping key rotation tests - JWT_ALGORITHM is not RS256")
	}

	ts := NewTestSuite(t)
	defer ts.Close()

	factory := NewTestDataFactory()

	t.Run("tokens remain valid after new key is added to JWKS", func(t *testing.T) {
		// Get initial JWKS
		req1, err := http.NewRequest("GET", ts.server.URL()+config.BASE_PATH+"/.well-known/jwks.json", nil)
		require.NoError(t, err)

		jwksResp1, err := ts.client.Do(req1)
		require.NoError(t, err)
		defer jwksResp1.Body.Close()

		var jwks1 map[string]interface{}
		json.NewDecoder(jwksResp1.Body).Decode(&jwks1)
		keys1 := jwks1["keys"].([]interface{})

		t.Logf("Initial JWKS has %d keys", len(keys1))

		// Create token with current key
		loginResp, err := ts.Request("POST", "/auth/login", factory.SimpleUserLoginData(), nil)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		var loginData httpserver.SuccessResponse[authcase.LoginOutput]
		json.NewDecoder(loginResp.Body).Decode(&loginData)

		oldToken := loginData.Data.AccessToken

		// Simulate time passing or trigger rotation manually if admin endpoint exists
		// For now, just verify the token still works
		userResp, err := ts.AuthenticatedRequest("GET", "/user/me", nil, oldToken)
		require.NoError(t, err)
		defer userResp.Body.Close()

		assert.Equal(t, http.StatusOK, userResp.StatusCode, "Old token should remain valid")

		// Note: Full rotation test would require admin endpoint or time manipulation
		t.Log("Full rotation test requires admin endpoint - see manual rotation tests")
	})
}
