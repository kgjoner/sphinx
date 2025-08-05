package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	oauthcase "github.com/kgjoner/sphinx/internal/domains/auth/cases/oauth/provider"
)

type AuthGateway struct {
	common.Pools
	common.Services
	mid common.Middlewares
}

func Raise(router chi.Router, pools common.Pools, services common.Services) {
	authgtw := &AuthGateway{
		pools,
		services,
		common.Middlewares{Pools: pools},
	}

	router.Route("/user", authgtw.userHandler)
	router.Route("/application", authgtw.applicationHandler)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", authgtw.login)
		r.Post("/external", authgtw.externalAuth)
		r.With(authgtw.mid.Authenticate).Post("/logout", authgtw.logout)
		r.With(authgtw.mid.Authenticate).Post("/refresh", authgtw.refresh)
	})

	router.Route("/oauth", func(r chi.Router) {
		r.With(authgtw.mid.Authenticate).Post("/authorize", authgtw.issueGrant)
		r.Post("/token", authgtw.exchangeGrant)
	})
}

// Login godoc
//
//	@Summary		Log user in
//	@Description	Exchange user credentials for root app auth tokens
//	@Router			/auth/login [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		authcase.LoginInput	true	"Credentials. Entry can be: email, phone, username or document"
//	@Success		200		{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) login(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddIp().
		AddHeader("user-agent", "device")

	var input authcase.LoginInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := authcase.Login{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// Logout godoc
//
//	@Summary		Log user out
//	@Description	Invalidate current session and their tokens
//	@Router			/auth/logout [post]
//	@Tags			Auth
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	normalizederr.NormalizedError
//	@Failure		401	{object}	normalizederr.NormalizedError
//	@Failure		500	{object}	normalizederr.NormalizedError
func (g AuthGateway) logout(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor()

	var input authcase.LogoutInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := authcase.Logout{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
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
//	@Param			payload	body		oauthcase.IssueGrantInput	true
//	@Success		200		{object}	presenter.Success[oauthcase.IssueGrantOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) issueGrant(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddActor()

	var input oauthcase.IssueGrantInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := oauthcase.IssueGrant{
		AuthRepo:  queries,
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
//	@Param			payload	body		auth.OAuthAuthenticateFields	true	"You must inform either client_secret or code_verifier"
//	@Success		200		{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) exchangeGrant(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddIp().
		AddHeader("user-agent", "device")

	var input oauthcase.ExchangeGrantInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := oauthcase.ExchangeGrant{
		AuthRepo:  queries,
		CacheRepo: g.CachePool.NewDAO(r.Context()),
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// Refresh godoc
//
//	@Summary		Issue new auth tokens
//	@Description	Get new auth tokens from a refresh one. This is invalidated.
//	@Router			/auth/refresh [post]
//	@Tags			Auth
//	@Security		BearerRefresh
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400	{object}	normalizederr.NormalizedError
//	@Failure		401	{object}	normalizederr.NormalizedError
//	@Failure		500	{object}	normalizederr.NormalizedError
func (g AuthGateway) refresh(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor()

	var input authcase.RefreshInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := authcase.Refresh{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// External Auth godoc
//
//	@Summary		Log user in via external provider
//	@Description	Use an external provider to authenticate user and generate root app auth tokens. It may create a new user or relate an existing one to the external provider if user had consented.
//	@Router			/auth/external [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		authcase.ExternalAuthInput	true	"You may inform email if creation is expected and provider does not provide it."
//	@Success		200		{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		403		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) externalAuth(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddIp().
		AddHeader("user-agent", "device").
		AddHeader("authorization", "authorizationHeader")

	var input authcase.ExternalAuthInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx common.BaseRepo) (any, error) {
		i := authcase.ExternalAuth{
			AuthRepo:    tx,
		}

		return i.Execute(input)
	})

	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
