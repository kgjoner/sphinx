package testserver

import (
	"net/http/httptest"

	"github.com/kgjoner/sphinx/internal/config"
	baserepo "github.com/kgjoner/sphinx/internal/repositories/base"
	"github.com/kgjoner/sphinx/internal/server"
)

// TestServer wraps the real server for testing
type TestServer struct {
	httpServer *httptest.Server
	realServer *server.Server
}

// New creates a test server using the real server implementation
func New() *TestServer {
	// Load config
	config.Must()
	config.Env.DATABASE_URL = "postgres://postgres:postgres@localhost:5433/sphinx_test?sslmode=disable"
	config.Env.REDIS_URL = "redis://localhost:6380/0"
	config.Env.HERMES.BASE_URL = "http://localhost:8082/v1"

	// Create the real server
	realServer := server.New().Setup()

	// Wrap it in httptest server
	httpServer := httptest.NewServer(realServer.Handler)

	return &TestServer{
		httpServer: httpServer,
		realServer: realServer,
	}
}

func (ts *TestServer) URL() string {
	return ts.httpServer.URL
}

func (ts *TestServer) Close() {
	ts.httpServer.Close()
}

// GetBasePool returns the database pool from the real server
func (ts *TestServer) GetBasePool() *baserepo.Pool {
	return ts.realServer.BasePool()
}

// TestServerBuilder provides a fluent API for building test servers
type TestServerBuilder struct {
	withConfig func()
}

func NewBuilder() *TestServerBuilder {
	return &TestServerBuilder{}
}

func (b *TestServerBuilder) WithConfig(configFn func()) *TestServerBuilder {
	b.withConfig = configFn
	return b
}

func (b *TestServerBuilder) Build() *TestServer {
	if b.withConfig != nil {
		b.withConfig()
	} else {
		config.Must()
	}

	return New()
}
