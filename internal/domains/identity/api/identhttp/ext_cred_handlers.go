package identhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	extcredcase "github.com/kgjoner/sphinx/internal/domains/identity/cases/extcred"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

func (g Gateway) externalCredentialHandler(r chi.Router) {
	r.Post("/external/{provider}", g.externalSignup)

	authedR := r.With(g.mid.Authenticate)
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
func (g Gateway) externalSignup(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("provider", "providerName").
		JSONBody().
		AddLanguages()

	var input extcredcase.SignUpInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx Repo) (any, error) {
		i := extcredcase.SignUp{
			IdentityRepo:     tx,
			AccessRepo:       tx,
			IdentityProvider: g.IdentityProvider,
			Hasher:           g.Services.Hasher,
			Mailer:           g.Services.Mailer,
		}
		return i.Execute(input)
	})
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
//		@Param			payload			body		extcredcase.AuthorizeExternalCredentialInput	true	"Parameters to authenticate with external identity provider."
//		@Success		200				{object}	presenter.Success[identity.ExternalCredentialView]
//		@Failure		400				{object}	apperr.AppError
//		@Failure		422				{object}	apperr.AppError
//		@Failure		500				{object}	apperr.AppError
func (g Gateway) authorizeExternalCredential(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("provider", "providerName").
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		JSONBody()

	var input extcredcase.AuthorizeExternalCredentialInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx Repo) (any, error) {
		i := extcredcase.AuthorizeExternalCredential{
			IdentityRepo:     tx,
			IdentityProvider: g.IdentityProvider,
		}
		return i.Execute(input)
	})
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
func (g Gateway) unauthorizeExternalCredential(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("provider", "providerName").
		ParseURLParam("subjectID", "providerSubjectID").
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input extcredcase.UnauthorizeExternalCredentialInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx Repo) (any, error) {
		i := extcredcase.UnauthorizeExternalCredential{
			IdentityRepo: tx,
		}
		return i.Execute(input)
	})
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}
