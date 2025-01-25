package sphinxup

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/repositories/cache"
	oauthcase "github.com/kgjoner/sphinx/pkg/sphinxup/cases/oauth"
)

type oAuthGateway struct {
	pool cache.Pool
	envs OAuthGatewayEnvs
}

type OAuthGatewayEnvs struct {
	SphinxApiBaseUrl    string
	SphinxClientBaseUrl string
	AppBaseUrl          string
	AppId               string
	AppSecret           string

	// Used to minimize security settings, allowing localhost to receive cookies, for example.
	// Ignore it in productions environments.
	Development bool
}

func RaiseOAuthGateway(router chi.Router, pool cache.Pool, envs OAuthGatewayEnvs) {
	oauthgtw := &oAuthGateway{
		pool,
		envs,
	}

	router.Route("/oauth", func(r chi.Router) {
		r.Get("/start", oauthgtw.start)
		r.Get("/callback", oauthgtw.callback)
		r.Post("/retrieve", oauthgtw.retrieve)
	})

	router.Get("/token-ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func (g oAuthGateway) start(w http.ResponseWriter, r *http.Request) {
	repo := g.pool.NewDAO(r.Context())
	i := oauthcase.StartOAuth{
		CacheRepo: repo,
	}

	origin := r.Header.Get("origin")
	output, cookie, err := i.Execute(oauthcase.StartOAuthInput{
		Origin:              origin,
		SphinxClientBaseUrl: g.envs.SphinxClientBaseUrl,
		AppBaseUrl:          g.envs.AppBaseUrl,
		AppId:               g.envs.AppId,
		Development:         g.envs.Development,
	})

	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	http.SetCookie(w, cookie)
	presenter.HttpSuccess(output, w, r)
}

func (g oAuthGateway) callback(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseQueryParam("code").
		ParseQueryParam("state")

	var input oauthcase.OAuthCallbackInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	input.AppId = g.envs.AppId
	input.AppSecret = g.envs.AppSecret
	input.SphinxApiBaseUrl = g.envs.SphinxApiBaseUrl

	repo := g.pool.NewDAO(r.Context())
	i := oauthcase.OAuthCallback{
		CacheRepo: repo,
	}

	origin, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	if origin == "" {
		origin = g.envs.AppBaseUrl
	}

	to := origin + "/token-ready"
	http.Redirect(w, r, to, http.StatusFound)
}

func (g oAuthGateway) retrieve(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddHeader("x-csrf-token", "csrfToken").
		ParseBody("state")

	var input oauthcase.RetrieveTokenInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	repo := g.pool.NewDAO(r.Context())
	i := oauthcase.RetrieveToken{
		CacheRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
