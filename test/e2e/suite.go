package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/test/testserver"
)

// TestSuite provides a structured approach to E2E testing
type TestSuite struct {
	server *testserver.TestServer
	client *http.Client
	t      *testing.T
}

func NewTestSuite(t *testing.T) *TestSuite {
	// Load test configuration
	config.Must()

	// Create test server with the actual server setup
	server := testserver.New()

	return &TestSuite{
		server: server,
		client: &http.Client{},
		t:      t,
	}
}

// Close cleans up the test server
func (ts *TestSuite) Close() {
	if ts.server != nil {
		ts.server.Close()
	}
}

// Helper method to make requests
func (ts *TestSuite) Request(method, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}

	req, err := http.NewRequest(method, ts.server.URL()+"/api"+endpoint, &reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return ts.client.Do(req)
}

// Helper method to make authenticated requests
func (ts *TestSuite) AuthenticatedRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return ts.Request(method, endpoint, body, headers)
}

// Helper method to make app-authenticated requests
func (ts *TestSuite) AuthenticatedAppRequest(method, endpoint string, body interface{}, appId string, appSecret string) (*http.Response, error) {
	token := appId + ":" + appSecret
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))
	headers := map[string]string{
		"Authorization": "Basic " + encodedToken,
	}
	return ts.Request(method, endpoint, body, headers)
}

// Deprecated: There is no need of x-app anymore. Helper method to make app-authenticated requests
func (ts *TestSuite) AppRequest(method, endpoint string, body interface{}, appId string) (*http.Response, error) {
	headers := map[string]string{
		"x-app": appId,
	}
	return ts.Request(method, endpoint, body, headers)
}
