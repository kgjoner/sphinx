package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/cornucopia/v2/repositories/cache"
	"github.com/kgjoner/cornucopia/v2/repositories/cache/redisdb"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/docs"
	"github.com/kgjoner/sphinx/internal/assets/img"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/config"
	authgtw "github.com/kgjoner/sphinx/internal/domains/auth/gateway"
	baserepo "github.com/kgjoner/sphinx/internal/repositories/base"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	basePool  *baserepo.Pool
	cachePool cache.Pool
	mailSvc   *hermes.Service
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

	mailSvc := hermes.New(config.Env.HERMES.BASE_URL, config.Env.HERMES.API_KEY)

	return &Server{
		basePool:  db,
		cachePool: rdb,
		mailSvc:   mailSvc,
	}
}

//	@title			Sphinx API
//	@version		{{ .Version }}
//	@description	An authentication and authorization server.

//	@contact.name	Kaio Rosa
//	@contact.url	https://dev.kgjoner.com.br
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
// @basePath	{{ .BasePath }}
func (s *Server) Setup() *Server {
	pools := common.Pools{
		BasePool:  s.basePool,
		CachePool: s.cachePool,
	}

	updateHermesStyle(s.mailSvc)
	services := common.Services{
		MailService: s.mailSvc,
	}

	baseR := chi.NewRouter()
	baseR.Use(realIP())
	baseR.Use(middleware.Timeout(60 * time.Second))
	baseR.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language", "User-Agent", "X-App", "X-Entry", "X-Target"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r := chi.NewRouter()
	baseR.Mount(config.BASE_PATH, r)

	// Api versioning
	r.Route("/api", func(r chi.Router) {
		r.Use(countRequestMetric())

		authgtw.Raise(r, pools, services)
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(config.Env.APP_VERSION))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Mount("/metrics", promhttp.Handler())

	// Root app files
	r.Route("/assets", func(r chi.Router) {
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
	docs.SwaggerInfo.Version = config.Env.APP_VERSION
	docs.SwaggerInfo.Host = config.Env.HOST
	docs.SwaggerInfo.Schemes = []string{config.Env.SCHEME}
	docs.SwaggerInfo.BasePath = config.BASE_PATH + "/api"

	r.Route("/docs", func(r chi.Router) {
		if len(config.Env.SWAGGER_AUTH) > 0 {
			r.Use(middleware.BasicAuth("Swagger", config.Env.SWAGGER_AUTH))
		}
		r.Get("/*", httpSwagger.Handler(
			httpSwagger.URL(config.BASE_PATH+"/docs/doc.json"),
		))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.BASE_PATH+"/docs/", http.StatusTemporaryRedirect)
	})

	s.Handler = baseR
	return s
}

func (s *Server) Start() {
	defer s.basePool.Close()
	defer s.cachePool.Close()

	server := &http.Server{
		Addr:    ":8080",
		Handler: s.Handler,
	}

	// Start the server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		fmt.Println("Server running at port 8080")
		serverErr <- server.ListenAndServe()
	}()

	// Start all jobs
	jobCtx, cancelJobs := context.WithCancel(context.Background())
	defer cancelJobs()
	s.runJobs(jobCtx)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Println("Server startup error: ", err)
		}
		return
	case <-quit:
		fmt.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Server forced to shutdown: ", err)
		}
		return
	}
}
