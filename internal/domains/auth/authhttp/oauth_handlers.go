package authhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

func (g gateway) oauthHandlers(r chi.Router) {
	r.With(g.Authenticate).Post("/authorize", g.issueGrant)
	r.Post("/token", g.exchangeGrant)
}

// Issue Grant godoc
//
//	@Summary		Request an authentication grant
//	@Description	Issue a new authentication grant for a third party application (OAuth 2.0)
//	@Router			/oauth/authorize [post]
//	@Tags			Auth
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		authcase.IssueGrantInput	true "OAUTH 2.0 parameters"
//	@Success		200		{object}	presenter.Success[authcase.IssueGrantOutput]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) issueGrant(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input authcase.IssueGrantInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.AuthPool.NewDAO(r.Context())
	i := authcase.IssueGrant{
		AuthRepo:  repo,
		CacheRepo: g.CachePool.NewDAO(r.Context()),
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// Exchange Grant godoc
//
//	@Summary		Exchange grant for tokens
//	@Description	Exchange authentication grant for third party application auth tokens
//	@Router			/oauth/token [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		auth.GrantCredentials	true	"You must inform either client_secret or code_verifier"
//	@Success		200		{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) exchangeGrant(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddIP().
		AddHeader("user-agent", "device")

	var input authcase.ExchangeGrantInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.AuthPool.NewDAO(r.Context())
	i := authcase.ExchangeGrant{
		AuthRepo:  repo,
		CacheRepo: g.CachePool.NewDAO(r.Context()),
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
