package authhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
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
//	@Success		200		{object}	httpserver.SuccessResponse[authcase.IssueGrantOutput]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) issueGrant(w http.ResponseWriter, r *http.Request) {
	var input authcase.IssueGrantInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	i := authcase.IssueGrant{
		AuthRepo:   g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection()),
		AccessRepo: g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection()),
		CacheRepo:  g.CachePool.NewStore(r.Context()),
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
//	@Success		200		{object}	httpserver.SuccessResponse[authcase.LoginOutput]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) exchangeGrant(w http.ResponseWriter, r *http.Request) {
	var input authcase.ExchangeGrantInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		IP(&input.IP).
		Header("user-agent", &input.Device)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := authcase.ExchangeGrant{
		AuthRepo:      repo,
		CacheRepo:     g.CachePool.NewStore(r.Context()),
		PwHasher:      g.PwHasher,
		DataHasher:    g.DataHasher,
		Challenger:    g.Challenger,
		TokenProvider: g.TokenProvider,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
}
