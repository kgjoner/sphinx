package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/kgjoner/sphinx/internal/pkg/identpvd"
)

var IdentityProviders = NewIdentityProvidersManager()

var Provider = struct {
	Name     string
}{
	Name: "testprovider",
}

var UnavailableProvider = struct {
	Name     string
}{
	Name: "unavailableprovider",
}

var Subject = struct {
	ID    string
	Email string
	Token string
}{
	ID:    "external-user-1",
	Email: "externaluser1@sphinx.test",
	Token: "valid-token-123",
}

func init() {
	IdentityProviders.AddProvider(Provider.Name)
	IdentityProviders.AddSubjectTo(Provider.Name, Subject.Token, Subject.ID, Subject.Email)

	IdentityProviders.AddProvider(UnavailableProvider.Name)
	IdentityProviders.AddSubjectTo(UnavailableProvider.Name, Subject.Token, Subject.ID, Subject.Email)

	unprov := IdentityProviders.Provider(UnavailableProvider.Name)
	unprov.SetError(http.StatusServiceUnavailable, "Service Unavailable")
}

/* =============================================================================
	Manager
============================================================================= */

// IdentityProvidersManager manages multiple mock providers for testing
type IdentityProvidersManager struct {
	providers map[string]*identityProvider
}

// NewIdentityProvidersManager creates a new manager for mock external auth providers
func NewIdentityProvidersManager() *IdentityProvidersManager {
	return &IdentityProvidersManager{
		providers: make(map[string]*identityProvider),
	}
}

// AddProvider adds a new mock provider with the given name
func (m *IdentityProvidersManager) AddProvider(name string) *identityProvider {
	provider := newMockIdentityProvider(name)
	m.providers[name] = provider
	return provider
}

// Provider returns a mock provider by name
func (m *IdentityProvidersManager) Provider(name string) *identityProvider {
	return m.providers[name]
}

// Config returns configurations for all mock providers as a JSON byte array
func (m *IdentityProvidersManager) Config() []byte {
	configs := make([]identpvd.Provider, 0, len(m.providers))
	for name, provider := range m.providers {
		configs = append(configs, provider.getConfig(name))
	}

	data, _ := json.Marshal(configs)
	return data
}

// Close shuts down all mock providers
func (m *IdentityProvidersManager) Close() {
	for _, provider := range m.providers {
		provider.close()
	}
}

// SetupTestProvider is a convenience function to set up a basic test provider
func (m *IdentityProvidersManager) AddSubjectTo(providerName, token, userID, email string) *identityProvider {
	provider := m.Provider(providerName)
	provider.addValidToken(token, &externalSubject{
		ID:    userID,
		Email: email,
		Name:  "Test User",
	})
	return provider
}

/* =============================================================================
	Provider
============================================================================= */

// identityProvider provides a mock external authentication provider
type identityProvider struct {
	Name              string
	server            *httptest.Server
	validTokens       map[string]*externalSubject
	expectedHeaders   map[string]string
	expectedParams    map[string]string
	responseData      map[string]interface{}
	shouldReturnError bool
	errorStatusCode   int
	errorMessage      string
}

// externalSubject represents a mock authenticated subject
type externalSubject struct {
	ID    string
	Email string
	Name  string
}

// newMockIdentityProvider creates a new mock external auth provider
func newMockIdentityProvider(name string) *identityProvider {
	mock := &identityProvider{
		Name:            name,
		validTokens:     make(map[string]*externalSubject),
		expectedHeaders: make(map[string]string),
		expectedParams:  make(map[string]string),
		responseData:    make(map[string]interface{}),
	}

	// Create the mock HTTP server
	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleAuthRequest))

	return mock
}

// Close shuts down the mock server
func (m *identityProvider) close() {
	if m.server != nil {
		m.server.Close()
	}
}

// URL returns the mock server URL
func (m *identityProvider) url() string {
	return m.server.URL
}

// AddValidToken adds a valid token with associated user data
func (m *identityProvider) addValidToken(token string, subject *externalSubject) {
	m.validTokens[token] = subject
}

// SetExpectedHeaders sets headers that should be present in the request
func (m *identityProvider) setExpectedHeaders(headers map[string]string) {
	m.expectedHeaders = headers
}

// SetExpectedParams sets query parameters that should be present in the request
func (m *identityProvider) setExpectedParams(params map[string]string) {
	m.expectedParams = params
}

// SetResponseData sets custom response data structure
func (m *identityProvider) setResponseData(data map[string]interface{}) {
	m.responseData = data
}

// SetError configures the mock to return an error response
func (m *identityProvider) SetError(statusCode int, message string) {
	m.shouldReturnError = true
	m.errorStatusCode = statusCode
	m.errorMessage = message
}

// ClearError removes error configuration
func (m *identityProvider) clearError() {
	m.shouldReturnError = false
	m.errorStatusCode = 0
	m.errorMessage = ""
}

// GetConfig returns a config.IdentityProvider configured to use this mock
func (m *identityProvider) getConfig(name string) identpvd.Provider {
	return identpvd.Provider{
		Name:           name,
		URL:            m.server.URL + "/auth",
		Method:         "GET",
		SubjectIDPath:  "user.id",
		EmailPath:      "user.email",
		DefaultHeaders: make(map[string]string),
		DefaultParams:  make(map[string]string),
		DefaultBody:    make(map[string]string),
	}
}

// handleAuthRequest handles incoming authentication requests
func (m *identityProvider) handleAuthRequest(w http.ResponseWriter, r *http.Request) {
	// Check if we should return an error
	if m.shouldReturnError {
		http.Error(w, m.errorMessage, m.errorStatusCode)
		return
	}

	// Extract authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	// Extract token from Bearer token
	token := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Check if token is valid
	subject, exists := m.validTokens[token]
	if !exists {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Validate expected headers
	for key, expectedValue := range m.expectedHeaders {
		if r.Header.Get(key) != expectedValue {
			http.Error(w, fmt.Sprintf("Missing or invalid header: %s", key), http.StatusBadRequest)
			return
		}
	}

	// Validate expected query parameters
	for key, expectedValue := range m.expectedParams {
		if r.URL.Query().Get(key) != expectedValue {
			http.Error(w, fmt.Sprintf("Missing or invalid parameter: %s", key), http.StatusBadRequest)
			return
		}
	}

	// Prepare response
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    subject.ID,
			"email": subject.Email,
			"name":  subject.Name,
		},
		"status": "authenticated",
	}

	// Override with custom response data if provided
	if len(m.responseData) > 0 {
		response = m.responseData
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
