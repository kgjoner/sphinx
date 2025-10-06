package testserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/cornucopia/v2/repositories/cache/memorydb"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/docs"
	"github.com/kgjoner/sphinx/internal/assets/img"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/config"
	authgtw "github.com/kgjoner/sphinx/internal/domains/auth/gateway"
	"github.com/kgjoner/sphinx/test/mocks"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type TestServer struct {
	httpServer      *httptest.Server
	mockRepoFactory *mocks.MockBasePool
	mockCachePool   *memorydb.Pool
	mockMailService *hermes.MockService
}

func New() *TestServer {
	// Initialize mocks
	mockRepoFactory := mocks.NewMockBasePool()
	mockCachePool, _ := memorydb.NewPool()
	mockMailService := hermes.NewMock()

	// Create pools and services using mocks
	pools := common.Pools{
		BasePool:  mockRepoFactory,
		CachePool: mockCachePool,
	}

	services := common.Services{
		MailService: mockMailService,
	}

	// Setup router like the original server
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language", "User-Agent", "X-App", "X-Entry", "X-Target"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Api versioning
	r.Route("/api", func(r chi.Router) {
		authgtw.Raise(r, pools, services)
	})

	// Root app files (copied from original server)
	r.Route("/asset", func(r chi.Router) {
		r.Get("/logo.svg", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Content-Length", fmt.Sprintf("%v", len(img.Logo)))
			w.Write(img.Logo)
		})
		r.Get("/style", func(w http.ResponseWriter, r *http.Request) {
			presenter.HTTPSuccess(style.Root, w, r)
		})
	})

	// Docs
	docs.SwaggerInfo.Host = "localhost:8080" // Use test host
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})

	// Create test server
	httpServer := httptest.NewServer(r)

	return &TestServer{
		httpServer:      httpServer,
		mockRepoFactory: mockRepoFactory,
		mockCachePool:   mockCachePool,
		mockMailService: mockMailService,
	}
}

func (ts *TestServer) URL() string {
	return ts.httpServer.URL
}

func (ts *TestServer) Close() {
	ts.httpServer.Close()
}

// Test helper methods
func (ts *TestServer) GetMockQueries() *mocks.MockQueries {
	return ts.mockRepoFactory.GetMockQueries()
}

func (ts *TestServer) GetMockCache() *memorydb.Pool {
	return ts.mockCachePool
}

func (ts *TestServer) GetMockMailService() *hermes.MockService {
	return ts.mockMailService
}

func (ts *TestServer) Reset() {
	ts.mockRepoFactory.GetMockQueries().Clear()
	ts.mockMailService.ClearMails()
}

// TestServerBuilder provides a fluent API for building test servers with specific configurations
type TestServerBuilder struct {
	withConfig func()
	setupMocks func(*TestServer)
}

func NewBuilder() *TestServerBuilder {
	return &TestServerBuilder{}
}

func (b *TestServerBuilder) WithConfig(configFn func()) *TestServerBuilder {
	b.withConfig = configFn
	return b
}

func (b *TestServerBuilder) WithMocks(setupFn func(*TestServer)) *TestServerBuilder {
	b.setupMocks = setupFn
	return b
}

func (b *TestServerBuilder) Build() *TestServer {
	// Load configuration if needed
	if b.withConfig != nil {
		b.withConfig()
	} else {
		// Load default test config
		config.Must()
	}

	server := New()

	// Setup mocks if provided
	if b.setupMocks != nil {
		b.setupMocks(server)
	}

	return server
}
