package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth/gateway"
	authrepo "github.com/kgjoner/sphinx/internal/domains/auth/repository"
	"github.com/kgjoner/sphinx/postgres"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	queries *psqlrepo.Queries
	Handler http.Handler
}

func New(pool *sql.DB) *Server {
	s := &Server{
		queries: psqlrepo.New(pool),
	}
	s.Setup()
	return s
}

func (s *Server) Setup() {
	repos := common.Repos{
		AuthRepo: authrepo.New(s.queries),
	}

	services := common.Services{
		MailService: hermes.New(config.Environment.HERMES.BASE_URL, config.Environment.HERMES.API_KEY),
	}

	r := chi.NewRouter()
	r.Use(addContextToReposAndServices(&repos))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "User-Agent", "X-Forwarded-For"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Api versioning
	r.Route("/v1", func(r chi.Router) {
		r.Use(countRequestMetric())

		authgtw.Raise(r, repos, services)
	})

	r.Mount("/metrics", promhttp.Handler())

	s.Handler = r
}

func (s *Server) Start() {
	fmt.Println("Server running at port 8080")
	http.ListenAndServe(":8080", s.Handler)
}

func addContextToReposAndServices(repos *common.Repos) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			type WithContext interface {
				AddContext(context.Context)
			}

			repos := []any{repos.AuthRepo}
			for _, repo := range repos {
				if ctxRepo, ok := repo.(WithContext); ok {
					ctxRepo.AddContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

var (
	RequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_request_count",
		Help: "The total number of requests",
	})
)

func countRequestMetric() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			RequestCounter.Inc()
			next.ServeHTTP(w, r)
		})
	}
}
