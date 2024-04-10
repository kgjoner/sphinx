package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/docs"
	"github.com/kgjoner/sphinx/internal/assets/img"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth/gateway"
	authrepo "github.com/kgjoner/sphinx/internal/domains/auth/repository"
	"github.com/kgjoner/sphinx/postgres"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/http-swagger/v2"
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

//	@title			Sphinx API
//	@version		0.1
//	@description	An authentication and authorization server.

//	@contact.name	Kaio Rosa
//	@contact.url	http://dev.kgjoner.com.br
//	@contact.email	dev@kgjoner.com.br

//	@securityDefinitions.apiKey	AppToken
//	@in							header
//	@name						Authorization
//	@description				Provide an identification of a valid registered Sphinx app

//	@securityDefinitions.apiKey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and a JWT Access token.

//	@securityDefinitions.apiKey	BearerRefresh
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and a JWT Refresh token.

// @host		{{ .Host }}
// @basePath	/v1
func (s *Server) Setup() {
	repos := common.RepoFactories{
		AuthRepo: authrepo.NewFactory(s.queries),
	}

	services := common.Services{
		MailService: hermes.New(config.Environment.HERMES.BASE_URL, config.Environment.HERMES.API_KEY),
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language", "User-Agent", "X-Forwarded-For"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Api versioning
	r.Route("/v1", func(r chi.Router) {
		r.Use(countRequestMetric())

		authgtw.Raise(r, repos, services)
	})

	r.Mount("/metrics", promhttp.Handler())

	// Root app files
	r.Route("/root", func(r chi.Router) {
		r.Get("/logo.svg", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Content-Length", fmt.Sprintf("%v", len(img.Logo)))
			w.Write(img.Logo)
		})
		r.Get("/style", func(w http.ResponseWriter, r *http.Request) {
			presenter.HttpSuccess(style.Root, w, r);
		})
	})

	// Docs
	docs.SwaggerInfo.Host = config.Environment.SWAGGER_HOST
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})

	s.Handler = r
}

func (s *Server) Start() {
	fmt.Println("Server running at port 8080")
	http.ListenAndServe(":8080", s.Handler)
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
