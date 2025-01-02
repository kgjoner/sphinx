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
	r.With(g.mid.AppId).Post("/", g.createAccount)
	r.With(g.mid.AppId).Post("/password/request", g.requestPasswordReset)
	r.With(g.mid.AppId).Patch("/{id}/password", g.resetPassword)
	
	r.With(g.mid.Authenticate, g.mid.Target).Get("/", g.getPrivateAccount)
	r.With(g.mid.Authenticate).Patch("/password", g.changePassword)

	r.With(g.mid.AuthenticateApp, g.mid.Target).Patch("/permission", g.editAccountPermissions)
	r.With(g.mid.AuthenticateApp, g.mid.Target).Get("/email", g.getAccountEmail)
	r.With(g.mid.AuthenticateApp).Get("/id", g.getAccountId)

	r.Get("/existence", g.checkEntryExistence)
	r.Patch("/{id}/verification", g.verifyAccount)
}

// CreateAccount godoc
//
//	@Summary		Create an account
//	@Description	Register a new account linked to source app, and send email validation code.
//	@Router			/account [post]
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			x-app			header		string						true	"Application ID"
//	@Param			accept-language	header		string						false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		auth.AccountCreationFields	true	"Email and password are mandatory."
//	@Success		200				{object}	presenter.Success[auth.Account]
//	@Failure		400				{object}	normalizederr.NormalizedError
//	@Failure		401				{object}	normalizederr.NormalizedError
//	@Failure		500				{object}	normalizederr.NormalizedError
func (g AuthGateway) createAccount(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(accountcase.CreateAccountInput{}.AccountCreationFields).JsonKeys()
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

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.CreateAccount{
		AuthRepo:    queries,
		CacheRepo:   g.CachePool.NewDAO(r.Context()),
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r, http.StatusCreated)
}

// CheckEntryExistence godoc
//
//	@Summary		Check whether an entry already is registered
//	@Description	Check whether email, username, phone or document has already been registered.
//	@Router			/account/existence [get]
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			x-entry	header		string	true	"Email, username, phone or document."
//	@Success		200		{object}	presenter.Success[bool]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) checkEntryExistence(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddHeader("X-Entry", "entry")

	var input accountcase.CheckEntryExistenceInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.CheckEntryExistence{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// GetPrivateAccounnt godoc
//
//	@Summary		Get account private data
//	@Description	Retrieve private data associated with logged account or target one, if x-target header is informed. The latter require special permission.
//	@Router			/account [get]
//	@Tags			Account
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header		string	false	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible."
//	@Success		200			{object}	presenter.Success[auth.AccountPrivateView]
//	@Failure		400			{object}	normalizederr.NormalizedError
//	@Failure		401			{object}	normalizederr.NormalizedError
//	@Failure		403			{object}	normalizederr.NormalizedError
//	@Failure		500			{object}	normalizederr.NormalizedError
func (g AuthGateway) getPrivateAccount(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor().
		AddTarget()

	var input accountcase.GetPrivateAccountInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.GetPrivateAccount{
		AuthRepo: queries,
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
//	@Router			/account/{id}/verification [patch]
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Account ID"
//	@Param			payload	body		accountcase.VerifyAccountInput	true	"Code kind must be email or phone."
//	@Success		200		{object}	presenter.Success[bool]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) verifyAccount(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody(structop.New(accountcase.VerifyAccountInput{}).JsonKeys()...).
		ParseUrlParam("id", "accountId")

	var input accountcase.VerifyAccountInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.VerifyAccount{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// ChangePassword godoc
//
//	@Summary		Update password of logged account.
//	@Description	Update password of logged account. It must provide the current password.
//	@Router			/account/password [patch]
//	@Tags			Account
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string							false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		accountcase.ChangePasswordInput	true	"Old password and new one."
//	@Success		200				{object}	presenter.Success[bool]
//	@Failure		400				{object}	normalizederr.NormalizedError
//	@Failure		401				{object}	normalizederr.NormalizedError
//	@Failure		500				{object}	normalizederr.NormalizedError
func (g AuthGateway) changePassword(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody(structop.New(accountcase.ChangePasswordInput{}).JsonKeys()...).
		AddActor().
		AddLanguages()

	var input accountcase.ChangePasswordInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.ChangePassword{
		AuthRepo:    queries,
		CacheRepo:   g.CachePool.NewDAO(r.Context()),
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// RequestPasswordReset godoc
//
//	@Summary		Request for a password reset.
//	@Description	Request for a password reset. An email is sent with instructions.
//	@Router			/account/password/request [post]
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			x-app			header		string						true	"Application ID"
//	@Param			accept-language	header		string									false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		accountcase.RequestPasswordResetInput	true	"Old password and new one."
//	@Success		200				{object}	presenter.Success[bool]
//	@Failure		400				{object}	normalizederr.NormalizedError
//	@Failure		500				{object}	normalizederr.NormalizedError
func (g AuthGateway) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody(structop.New(accountcase.RequestPasswordResetInput{}).JsonKeys()...).
		AddApplication().
		AddLanguages()

	var input accountcase.RequestPasswordResetInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.RequestPasswordReset{
		AuthRepo:    queries,
		CacheRepo:   g.CachePool.NewDAO(r.Context()),
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// ResetPassword godoc
//
//	@Summary		Update password with a reset code.
//	@Description	Update password with a reset code.
//	@Router			/account/{id}/password [post]
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			x-app			header		string						true	"Application ID"
//	@Param			accept-language	header		string							false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			id				path		string							true	"Account ID"
//	@Param			payload			body		accountcase.ResetPasswordInput	true	"Old password and new one."
//	@Success		200				{object}	presenter.Success[bool]
//	@Failure		400				{object}	normalizederr.NormalizedError
//	@Failure		500				{object}	normalizederr.NormalizedError
func (g AuthGateway) resetPassword(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody(structop.New(accountcase.ResetPasswordInput{}).JsonKeys()...).
		ParseUrlParam("id", "accountId").
		AddApplication().
		AddLanguages()

	var input accountcase.ResetPasswordInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.ResetPassword{
		AuthRepo:    queries,
		CacheRepo:   g.CachePool.NewDAO(r.Context()),
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// EditAccountPermissions godoc
//
//	@Summary		Add or remove roles and grantings
//	@Description	Add or remove roles and/or grantings of the target account.
//	@Router			/account/permission [patch]
//	@Tags			Account
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header		string									true	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible."
//	@Param			payload		body		accountcase.EditAccountPermissionsInput	true	"At least one of roles and grantings must be defined"
//	@Success		200			{object}	presenter.Success[bool]
//	@Failure		400			{object}	normalizederr.NormalizedError
//	@Failure		401			{object}	normalizederr.NormalizedError
//	@Failure		403			{object}	normalizederr.NormalizedError
//	@Failure		500			{object}	normalizederr.NormalizedError
func (g AuthGateway) editAccountPermissions(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(accountcase.EditAccountPermissionsInput{}).JsonKeys()
	c := controller.New(r).
		ParseBody(bodyKeys...).
		AddApplication().
		AddTarget()

	var input accountcase.EditAccountPermissionsInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.EditAccountPermissions{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// GetAccounntId godoc
//
//	@Summary		Get account id
//	@Description	Retrieve account id by its entry. Return nil if entry does not exist.
//	@Router			/account/id [get]
//	@Tags			Account
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			x-entry	header		string	true	"Email, username, phone or document."
//	@Success		200			{object}	presenter.Success[string]
//	@Failure		400			{object}	normalizederr.NormalizedError
//	@Failure		401			{object}	normalizederr.NormalizedError
//	@Failure		403			{object}	normalizederr.NormalizedError
//	@Failure		500			{object}	normalizederr.NormalizedError
func (g AuthGateway) getAccountId(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddHeader("X-Entry", "entry")

	var input accountcase.GetAccountIdInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.GetAccountId{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// GetAccountEmail godoc
//
//	@Summary		Get target account email
//	@Description	Retrieve email of the target account.
//	@Router			/account/email [get]
//	@Tags			Account
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header		string									true	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible."
//	@Success		200			{object}	presenter.Success[string]
//	@Failure		400			{object}	normalizederr.NormalizedError
//	@Failure		401			{object}	normalizederr.NormalizedError
//	@Failure		403			{object}	normalizederr.NormalizedError
//	@Failure		500			{object}	normalizederr.NormalizedError
func (g AuthGateway) getAccountEmail(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddTarget()

	var input accountcase.GetAccountEmailInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := accountcase.GetAccountEmail{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
