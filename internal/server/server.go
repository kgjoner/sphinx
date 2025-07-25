package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/cornucopia/repositories/cache/redisdb"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/docs"
	"github.com/kgjoner/sphinx/internal/assets/img"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/config"
	authgtw "github.com/kgjoner/sphinx/internal/domains/auth/gateway"
	baserepo "github.com/kgjoner/sphinx/internal/repositories/base"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	basePool  *baserepo.Pool
	cachePool cache.Pool
	Handler   http.Handler
}

func New() *Server {
	db, err := baserepo.NewPool(config.Env.DATABASE_URL)
	if err != nil {
		log.Fatalln(err)
	}

	rdb, err := redisdb.NewPool(config.Env.REDIS_URL)
	if err != nil {
		log.Fatalln(err)
	}

	return &Server{
		basePool:  db,
		cachePool: rdb,
	}
}

//	@title			Sphinx API
//	@version		0.1
//	@description	An authentication and authorization server.

//	@contact.name	Kaio Rosa
//	@contact.url	http://dev.kgjoner.com.br
//	@contact.email	dev@kgjoner.com.br

//	@securityDefinitions.basic	BasicApp
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
func (s *Server) Setup() *Server {
	pools := common.Pools{
		BasePool:  s.basePool,
		CachePool: s.cachePool,
	}

	services := common.Services{
		MailService: hermes.New(config.Env.HERMES.BASE_URL, config.Env.HERMES.API_KEY, hermes.Options{
			PrimaryColor:      style.Root.Colors.PrimaryPure,
			PrimaryHoverColor: style.Root.Colors.PrimaryDark,
			Header: struct {
				Logo      string       "json:\"logo\""
				Title     string       "json:\"title\""
				Style     template.CSS "json:\"style\""
				LogoStyle template.CSS "json:\"logoStyle\""
			}{
				Logo:  config.Env.HOST + "/root/logo.svg",
				Title: config.Env.APP_NAME,
				Style: template.CSS(fmt.Sprintf("background-color: %v;", style.Root.Colors.BackgroundLight)),
			},
			Footer: struct {
				Text  string       "json:\"text\""
				Style template.CSS "json:\"style\""
			}{
				Style: template.CSS(fmt.Sprintf("background-color: %v;", style.Root.Colors.BackgroundDark)),
			},
			Alias: struct {
				Address htypes.Email "json:\"address\""
				Name    string       "json:\"name\""
			}{
				Name: config.Env.APP_NAME,
			},
		}),
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language", "User-Agent", "X-App", "X-Entry", "X-Target"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Api versioning
	r.Route("/v1", func(r chi.Router) {
		r.Use(countRequestMetric())

		authgtw.Raise(r, pools, services)
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
			presenter.HttpSuccess(style.Root, w, r)
		})
	})

	// Docs
	docs.SwaggerInfo.Host = config.Env.HOST
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})

	s.Handler = r
	return s
}

func (s *Server) Start() {
	defer s.basePool.Close()
	defer s.cachePool.Close()

	server := &http.Server{
		Addr:    ":8080",
		Handler: s.Handler,
	}

	go func() {
		fmt.Println("Server running at port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server startup error:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
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
