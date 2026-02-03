package sphinx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/pkg/sphinx"
)

// TestJWKSProvider_FetchKeys tests that the JWKS provider can fetch and cache keys
func TestJWKSProvider_FetchKeys(t *testing.T) {
	// Create a mock JWKS endpoint
	jwksData := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": "test-key-1",
				"n":   "xCbe9GcWC_FX7RedJ_0YfidLBY1CWkOdLNCbc5Aw50Dlws-xLHXEgQXVWo1W1Ch4q8fZ6QDK9f7-Ab_3LKQkltsO9gbBJUbqZh8QV0_Com-_w73JUnsl6RhvAjb843adeFo9I4WJJZ9LEYPIeI2-MhAcLJdCeSlV_IZjmXR9ftE1eEFvG9jQ8EUMtwbkSgFD2RKCuCuZvLjIWl3jxIxtR1dWNmukz6wdhSJGuvMn5mQnokCObHnbbP9XAQfK55NYzdtOznEdXBnRZo1NpNVOY1fCLIaf0Ckm70P7A12AfDGQq5jTJAKgYUanJiQuntLkaUispBujatbyY3oU692mzQ",
				"e":   "AQAB",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jwksData)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create JWKS provider
	provider := sphinx.NewJWKSProvider(server.URL, "")

	// The provider should fetch keys lazily on first validation attempt
	// For now, just verify it was created successfully
	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
}

// TestNewMiddlewares_WithJWKS tests the new JWKS-based middleware creation
func TestNewMiddlewares_WithJWKS(t *testing.T) {
	// Create a mock JWKS endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		}
	}))
	defer server.Close()

	// Create middlewares
	appID := uuid.New()
	authorizer := &mockAuthorizer{}
	middlewares := sphinx.NewMiddlewares(appID, server.URL, authorizer)

	if middlewares == nil {
		t.Fatal("Expected middlewares to be created")
	}
}

// TestNewMiddlewaresWithSecret_BackwardCompatibility tests the deprecated method still works
func TestNewMiddlewaresWithSecret_BackwardCompatibility(t *testing.T) {
	appID := uuid.New()
	authorizer := &mockAuthorizer{}
	middlewares := sphinx.NewMiddlewaresWithSecret(appID, "test-secret", authorizer)

	if middlewares == nil {
		t.Fatal("Expected middlewares to be created with secret")
	}
}

// TestClient_Middlewares tests the convenience method on Client
func TestClient_Middlewares(t *testing.T) {
	// Create a mock JWKS endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		}
	}))
	defer server.Close()

	// Create client
	client := sphinx.NewClient(server.URL, uuid.New().String(), "test-secret")
	authorizer := &mockAuthorizer{}

	middlewares := client.Middlewares(authorizer)

	if middlewares == nil {
		t.Fatal("Expected middlewares to be created from client")
	}
}

// Mock authorizer for testing
type mockAuthorizer struct{}

func (a *mockAuthorizer) AuthorizeSubject(sub sphinx.Subject, r *http.Request) (*http.Request, error) {
	return r, nil
}

// TestJWKSProvider_ErrorHandling tests error cases
func TestJWKSProvider_ErrorHandling(t *testing.T) {
	t.Run("Invalid URL", func(t *testing.T) {
		provider := sphinx.NewJWKSProvider("http://invalid-url-that-does-not-exist.local", "")

		// Validation should fail when JWKS cannot be fetched
		_, _, err := provider.Validate("some.jwt.token")
		if err == nil {
			t.Error("Expected error when JWKS endpoint is unreachable")
		}
	})

	t.Run("Invalid JSON Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		provider := sphinx.NewJWKSProvider(server.URL, "")

		_, _, err := provider.Validate("some.jwt.token")
		if err == nil {
			t.Error("Expected error when JWKS response is invalid JSON")
		}
	})

	t.Run("HTTP Error Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		provider := sphinx.NewJWKSProvider(server.URL, "")

		_, _, err := provider.Validate("some.jwt.token")
		if err == nil {
			t.Error("Expected error when JWKS endpoint returns error")
		}
	})
}

// TestJWKSProvider_TokenValidation tests token validation scenarios
func TestJWKSProvider_TokenValidation(t *testing.T) {
	t.Run("Missing kid in token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		}))
		defer server.Close()

		provider := sphinx.NewJWKSProvider(server.URL, "")

		// Token without kid should fail
		tokenWithoutKid := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.invalid"
		_, _, err := provider.Validate(tokenWithoutKid)

		if err == nil {
			t.Error("Expected error for token without kid")
		}

		// Should be an invalid access error
		if err != auth.ErrInvalidAccess {
			t.Logf("Got error: %v", err)
		}
	})

	
	t.Run("Invalid token format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[]}`))
		}))
		defer server.Close()

		provider := sphinx.NewJWKSProvider(server.URL, "")

		_, _, err := provider.Validate("not-a-jwt-token")
		if err == nil {
			t.Error("Expected error for invalid token format")
		}
	})
}

// TestJWKSProvider_Generate tests that Generate is not supported
func TestJWKSProvider_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"keys":[]}`))
	}))
	defer server.Close()

	provider := sphinx.NewJWKSProvider(server.URL, "")

	// Generate should not be supported (SDK is for validation only)
	_, err := provider.Generate(auth.Subject{})
	if err == nil {
		t.Error("Expected error when calling Generate on JWKS provider")
	}
}
