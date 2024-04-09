package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/structop"
	accountcase "github.com/kgjoner/sphinx/internal/domains/auth/cases/account"
)

func (g AuthGateway) accountHandler(r chi.Router) {
	r.With(g.mid.AppToken).Post("/", g.createAccount)

	r.With(g.mid.Authenticate).Get("/self", g.getPrivateAccount)
	r.With(g.mid.Authenticate).Post("/{id}/verification", g.verifyAccount)
	r.With(g.mid.Authenticate).Get("/{entry}", g.getPrivateAccount)
	r.With(g.mid.Authenticate).Patch("/{entry}/permission", g.editAccountPermissions)
}

// CreateAccount godoc
//
//	@Summary		Create an account
//	@Description	Register a new account linked to source app, and send email validation code.
//	@Router			/account [post]
//	@Tags			Account
//	@Security		AppToken
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string						false	"pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		auth.AccountCreationFields	true	"Email and password are mandatory."
//	@Success		200				{object}	presenter.Success[auth.Account]
//	@Failure		400				{object}	normalizederr.NormalizedError
//	@Failure		401				{object}	normalizederr.NormalizedError
//	@Failure		500				{object}	normalizederr.NormalizedError
func (g AuthGateway) createAccount(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(accountcase.CreateAccountInput{}.AccountCreationFields).Keys()
	c := controller.New(r).
		ParseBody(bodyKeys...).
		AddApplication().
		AddLanguages()

	var input accountcase.CreateAccountInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := accountcase.CreateAccount{
		AuthRepo:    g.AuthRepo,
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r, http.StatusCreated)
}

// GetPrivateAccounnt godoc
//
//	@Summary		Get account private data
//	@Description	Retrieve private data associated with logged account (or target one, if high user)
//	@Router			/account/self [get]
//	@Router			/account/{entry} [get]
//	@Tags			Account
//	@Security		AppToken
//	@Accept			json
//	@Produce		json
//	@Param			entry	path		string	true	"Email, username, phone or document"
//	@Success		200		{object}	presenter.Success[auth.AccountPrivateView]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		403		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) getPrivateAccount(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor().
		ParseUrlParam("entry")

	var input accountcase.GetPrivateAccountInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := accountcase.GetPrivateAccount{
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// VerifyAccounnt godoc
//
//	@Summary		Verify account data
//	@Description	Verify email or phone of target account
//	@Router			/account/{id}/verification [post]
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		accountcase.VerifyAccountInput	true	"Code kind must be email or phone."
//	@Success		200		{object}	presenter.Success[bool]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) verifyAccount(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseUrlParam("id", "accountId").
		ParseBody("code", "codeKind")

	var input accountcase.VerifyAccountInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := accountcase.VerifyAccount{
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// EditAccounntPermissions godoc
//
//	@Summary		Add or remove roles and grantings
//	@Description	Add or remove roles and/or grantings of the target account. Must be a high user.
//	@Router			/account/{entry}/permission [patch]
//	@Tags			Account
//	@Security		AppToken
//	@Accept			json
//	@Produce		json
//	@Param			entry	path		string									true	"Email, username, phone or document"
//	@Param			payload	body		accountcase.EditAccountPermissionsInput	true	"At least one of roles and grantings must be defined"
//	@Success		200		{object}	presenter.Success[auth.AccountPrivateView]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		403		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) editAccountPermissions(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(accountcase.EditAccountPermissionsInput{}).Keys()
	c := controller.New(r).
		ParseBody(bodyKeys...).
		ParseUrlParam("entry", "targetAccountEntry").
		AddActor()

	var input accountcase.EditAccountPermissionsInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := accountcase.EditAccountPermissions{
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
