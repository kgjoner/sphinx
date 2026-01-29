package identhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/identity/identcase"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

func (g gateway) externalCredentialHandler(r chi.Router) {
	r.Post("/external/{provider}", g.externalSignup)

	authedR := r.With(g.Authenticate)
	authedR.Put("/me/external/{provider}", g.authorizeExternalCredential)
	authedR.Delete("/me/external/{provider}/{subjectID}", g.unauthorizeExternalCredential)
}

// ExternalSignUp godoc
//
//		@Summary		Sign up a new user with external identity provider
//		@Description	Verify external identity provider, and use it to register a new user linked to root app, and send email validation code.
//		@Router			/user/external/{provider} [post]
//		@Tags			User
//		@Accept			json
//		@Produce		json
//		@Param			accept-language	header		string					false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	 @Param			provider		path		string					true	"External identity provider name."
//		@Param			payload			body		shared.IdentityProviderInput	true	"Parameters to authenticate with external identity provider."
//		@Success		201				{object}	presenter.Success[identity.UserLeanView]
//		@Failure		400				{object}	apperr.AppError
//		@Failure		409				{object}	apperr.AppError
//		@Failure		422				{object}	apperr.AppError
//		@Failure		500				{object}	apperr.AppError
func (g gateway) externalSignup(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("provider", "providerName").
		JSONBody().
		AddLanguages()

	var input identcase.ExternalSignUpInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	identRepo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	accessRepo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.ExternalSignUp{
		IdentityRepo:     identRepo,
		AccessRepo:       accessRepo,
		IdentityProvider: g.IdentityProvider,
		PwHasher:         g.PwHasher,
		Mailer:           g.Mailer,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusCreated)
}

// AuthorizeExternalCredential godoc
//
//		@Summary		Link an external credential to the authenticated user
//		@Description	Authorize and link an external credential from an identity provider to the authenticated user's account.
//		@Router			/user/me/external/{provider} [put]
//		@Tags			User
//	@Security		Bearer
//		@Accept			json
//		@Produce		json
//	 @Param			provider		path		string					true	"External identity provider name."
//		@Param			payload			body		identcase.AuthorizeExternalCredentialInput	true	"Parameters to authenticate with external identity provider."
//		@Success		200				{object}	presenter.Success[identity.ExternalCredentialView]
//		@Failure		400				{object}	apperr.AppError
//		@Failure		422				{object}	apperr.AppError
//		@Failure		500				{object}	apperr.AppError
func (g gateway) authorizeExternalCredential(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("provider", "providerName").
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		JSONBody()

	var input identcase.AuthorizeExternalCredentialInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.AuthorizeExternalCredential{
		IdentityRepo:     repo,
		IdentityProvider: g.IdentityProvider,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// UnauthorizeExternalCredential godoc
//
//		@Summary		Unlink an external credential from the authenticated user
//		@Description	Unauthorize and unlink an external credential from an identity provider from the authenticated user's account.
//		@Router			/user/me/external/{provider}/{subjectID} [delete]
//		@Tags			User
//	@Security		Bearer
//		@Produce		json
//	 @Param			provider		path		string					true	"External identity provider name."
//	 @Param			subjectID		path		string					true	"External identity provider subject ID."
//		@Success		204
//		@Failure		400				{object}	apperr.AppError
//		@Failure		422				{object}	apperr.AppError
//		@Failure		500				{object}	apperr.AppError
func (g gateway) unauthorizeExternalCredential(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("provider", "providerName").
		ParseURLParam("subjectID", "providerSubjectID").
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input identcase.UnauthorizeExternalCredentialInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.UnauthorizeExternalCredential{
		IdentityRepo: repo,
	}
	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}
