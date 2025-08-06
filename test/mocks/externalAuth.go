package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/kgjoner/sphinx/internal/config"
)

// MockExternalAuthProvider provides a mock external authentication provider
type MockExternalAuthProvider struct {
	Name              string
	server            *httptest.Server
	validTokens       map[string]*MockAuthSubject
	expectedHeaders   map[string]string
	expectedParams    map[string]string
	responseData      map[string]interface{}
	shouldReturnError bool
	errorStatusCode   int
	errorMessage      string
}

// MockAuthSubject represents a mock authenticated subject
type MockAuthSubject struct {
	ID    string
	Email string
	Name  string
}

// NewMockExternalAuthProvider creates a new mock external auth provider
func NewMockExternalAuthProvider(name string) *MockExternalAuthProvider {
	mock := &MockExternalAuthProvider{
		Name:            name,
		validTokens:     make(map[string]*MockAuthSubject),
		expectedHeaders: make(map[string]string),
		expectedParams:  make(map[string]string),
		responseData:    make(map[string]interface{}),
	}

	// Create the mock HTTP server
	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleAuthRequest))

	return mock
}

// Close shuts down the mock server
func (m *MockExternalAuthProvider) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// URL returns the mock server URL
func (m *MockExternalAuthProvider) URL() string {
	return m.server.URL
}

// AddValidToken adds a valid token with associated user data
func (m *MockExternalAuthProvider) AddValidToken(token string, subject *MockAuthSubject) {
	m.validTokens[token] = subject
}

// SetExpectedHeaders sets headers that should be present in the request
func (m *MockExternalAuthProvider) SetExpectedHeaders(headers map[string]string) {
	m.expectedHeaders = headers
}

// SetExpectedParams sets query parameters that should be present in the request
func (m *MockExternalAuthProvider) SetExpectedParams(params map[string]string) {
	m.expectedParams = params
}

// SetResponseData sets custom response data structure
func (m *MockExternalAuthProvider) SetResponseData(data map[string]interface{}) {
	m.responseData = data
}

// SetError configures the mock to return an error response
func (m *MockExternalAuthProvider) SetError(statusCode int, message string) {
	m.shouldReturnError = true
	m.errorStatusCode = statusCode
	m.errorMessage = message
}

// ClearError removes error configuration
func (m *MockExternalAuthProvider) ClearError() {
	m.shouldReturnError = false
	m.errorStatusCode = 0
	m.errorMessage = ""
}

// GetConfig returns a config.ExternalAuthProvider configured to use this mock
func (m *MockExternalAuthProvider) GetConfig(name string) config.ExternalAuthProvider {
	return config.ExternalAuthProvider{
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
func (m *MockExternalAuthProvider) handleAuthRequest(w http.ResponseWriter, r *http.Request) {
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

// MockExternalAuthManager manages multiple mock providers for testing
type MockExternalAuthManager struct {
	providers map[string]*MockExternalAuthProvider
}

// NewMockExternalAuthManager creates a new manager for mock external auth providers
func NewMockExternalAuthManager() *MockExternalAuthManager {
	return &MockExternalAuthManager{
		providers: make(map[string]*MockExternalAuthProvider),
	}
}

// AddProvider adds a new mock provider with the given name
func (m *MockExternalAuthManager) AddProvider(name string) *MockExternalAuthProvider {
	provider := NewMockExternalAuthProvider(name)
	m.providers[name] = provider
	return provider
}

// GetProvider returns a mock provider by name
func (m *MockExternalAuthManager) GetProvider(name string) *MockExternalAuthProvider {
	return m.providers[name]
}

// GetConfigs returns config.ExternalAuthProvider configurations for all mock providers
func (m *MockExternalAuthManager) GetConfigs() []config.ExternalAuthProvider {
	configs := make([]config.ExternalAuthProvider, 0, len(m.providers))
	for name, provider := range m.providers {
		configs = append(configs, provider.GetConfig(name))
	}
	return configs
}

// Close shuts down all mock providers
func (m *MockExternalAuthManager) Close() {
	for _, provider := range m.providers {
		provider.Close()
	}
}

// SetupTestProvider is a convenience function to set up a basic test provider
func (m *MockExternalAuthManager) SetupTestProvider(name, token, userID, email string) *MockExternalAuthProvider {
	provider := m.AddProvider(name)
	provider.AddValidToken(token, &MockAuthSubject{
		ID:    userID,
		Email: email,
		Name:  "Test User",
	})
	return provider
}
