package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common"
	usercase "github.com/kgjoner/sphinx/internal/domains/auth/cases/user"
)

func (g AuthGateway) userHandler(r chi.Router) {
	r.Post("/", g.createUser)
	r.Get("/existence", g.checkEntryExistence)
	r.Post("/request-password", g.requestPasswordReset)
	r.Post("/{id}/password", g.resetPassword)
	r.Post("/{id}/{field}/verification", g.verifyUser)

	r.With(g.mid.Authenticate, g.mid.Target).Get("/me", g.getMe)
	r.With(g.mid.Authenticate, g.mid.Target).Patch("/me", g.updateMyExtraData)
	r.With(g.mid.Authenticate, g.mid.Target).Post("/me/password", g.changeMyPassword)
	r.With(g.mid.Authenticate, g.mid.Target).Post("/me/{field}", g.updateMyUniqueFields)
	r.With(g.mid.Authenticate, g.mid.Target).Delete("/me/{field}/verification", g.cancelMyPendingField)

	r.With(g.mid.Authenticate, g.mid.Target).Get("/{id}", g.getPrivateUser)
	r.With(g.mid.Authenticate, g.mid.Target).Patch("/{id}", g.updateExtraData)
	r.With(g.mid.Authenticate, g.mid.Target).Post("/{id}/{field}", g.updateUniqueFields)
	r.With(g.mid.AuthenticateApp, g.mid.Target).Patch("/{id}/permission", g.editUserPermissions)
	r.With(g.mid.AuthenticateApp, g.mid.Target).Get("/{id}/email", g.getUserEmail)

	// Utility endpoint to get user ID by its entry. Passed on X-Entry header.
	r.With(g.mid.AuthenticateApp).Get("/id", g.getUserID)
}

// CreateUser godoc
//
//	@Summary		Create an user
//	@Description	Register a new user linked to root app, and send email validation code.
//	@Router			/user [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string					false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		auth.UserCreationFields	true	"Email and password are mandatory."
//	@Success		201				{object}	presenter.Success[auth.UserView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g AuthGateway) createUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddLanguages()

	var input usercase.CreateUserInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx common.BaseRepo) (any, error) {
		i := usercase.CreateUser{
			AuthRepo:    tx,
			MailService: g.MailService,
		}

		return i.Execute(input)
	})

	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusCreated)
}

// CheckEntryExistence godoc
//
//	@Summary		Check whether an entry already is registered
//	@Description	Check whether email, username, phone or document has already been registered.
//	@Router			/user/existence [get]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			x-entry	header		string	true	"Email, username, phone or document."
//	@Success		200		{object}	presenter.Success[bool]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g AuthGateway) checkEntryExistence(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddHeader("X-Entry", "entry")

	var input usercase.CheckEntryExistenceInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.CheckEntryExistence{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// GetMe godoc
//
//	@Summary		Get authenticated user data
//	@Description	Retrieve private data associated with logged user
//	@Router			/user/me [get]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) getMe(w http.ResponseWriter, r *http.Request) {
	g.getPrivateUser(w, r)
}

// GetPrivateUser godoc
//
//	@Summary		Get user private data
//	@Description	Retrieve private data associated with a target user. Require special permission.
//	@Router			/user/{id} [get]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) getPrivateUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(common.ActorCtxKey, "actor").
		AddFromContext(common.TargetCtxKey, "target")

	var input usercase.GetPrivateUserInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.GetPrivateUser{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// VerifyUser godoc
//
//	@Summary		Verify user data
//	@Description	Verify email or phone of target user
//	@Router			/user/{id}/{field}/verification [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"User ID"
//	@Param			field	path	string						true	"Verification field (email or phone)"
//	@Param			payload	body	usercase.VerifyUserInput	true	"Code is required."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) verifyUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("id", "userID").
		ParseURLParam("field", "verificationKind")

	var input usercase.VerifyUserInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.VerifyUser{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// ChangeMyPassword godoc
//
//	@Summary		Update password of authenticated user.
//	@Description	Update password of authenticated user. It must provide the current password.
//	@Router			/user/me/password [post]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header	string							false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body	usercase.ChangePasswordInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) changeMyPassword(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(common.ActorCtxKey, "actor").
		AddLanguages()

	var input usercase.ChangePasswordInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx common.BaseRepo) (any, error) {
		i := usercase.ChangePassword{
			AuthRepo:    tx,
			MailService: g.MailService,
		}

		return i.Execute(input)
	})
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// RequestPasswordReset godoc
//
//	@Summary		Request for a password reset.
//	@Description	Request for a password reset. An email is sent with instructions.
//	@Router			/user/request-password [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header	string								false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body	usercase.RequestPasswordResetInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddLanguages()

	var input usercase.RequestPasswordResetInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.RequestPasswordReset{
		AuthRepo:    queries,
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// ResetPassword godoc
//
//	@Summary		Update password with a reset code.
//	@Description	Update password with a reset code.
//	@Router			/user/{id}/password [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header	string						false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			id				path	string						true	"User ID"
//	@Param			payload			body	usercase.ResetPasswordInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) resetPassword(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("id", "userID").
		AddLanguages()

	var input usercase.ResetPasswordInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx common.BaseRepo) (any, error) {
		i := usercase.ResetPassword{
			AuthRepo:    tx,
			MailService: g.MailService,
		}

		return i.Execute(input)
	})
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// EditUserPermissions godoc
//
//	@Summary		Add or remove roles
//	@Description	Add or remove roles of the target user. Must be used by authenticated applications with proper permissions.
//	@Router			/user/{id}/permission [patch]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string								true	"User ID"
//	@Param			payload	body	usercase.EditUserPermissionsInput	true	"At least one of roles and grantings must be defined"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) editUserPermissions(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(common.ApplicationCtxKey, "application").
		AddFromContext(common.TargetCtxKey, "target")

	var input usercase.EditUserPermissionsInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.EditUserPermissions{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// UpdateMyExtraData godoc
//
//	@Summary		Update extra data of authenticated user
//	@Description	Update non unique data like name, surname and address of authenticated user
//	@Router			/user/me [patch]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		auth.ExtraData	true	"At least one data must be defined.""
//	@Success		200		{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		422		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g AuthGateway) updateMyExtraData(w http.ResponseWriter, r *http.Request) {
	g.updateExtraData(w, r)
}

// UpdateExtraData godoc
//
//	@Summary		Update extra data of target user
//	@Description	Update non unique data like name, surname and address of target user
//	@Router			/user/{id} [patch]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"User ID"
//	@Param			payload	body		auth.ExtraData	true	"At least one data must be defined.""
//	@Success		200		{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		422		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g AuthGateway) updateExtraData(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(common.ActorCtxKey, "actor").
		AddFromContext(common.TargetCtxKey, "target")

	var input usercase.UpdateExtraDataInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.UpdateExtraData{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// UpdateMyUniqueFields godoc
//
//	@Summary		Update an unique field of authenticated user
//	@Description	Update one unique data of authenticated user: email, phone, username or document. Email and phone updates require verification.
//	@Router			/user/me/{field} [post]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string								false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			field			path		string								true	"Must be: email, phone, username or document"
//	@Param			payload			body		usercase.UpdateUniqueFieldsInput	true	"Value is required."
//	@Success		200				{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g AuthGateway) updateMyUniqueFields(w http.ResponseWriter, r *http.Request) {
	g.updateUniqueFields(w, r)
}

// UpdateUniqueFields godoc
//
//	@Summary		Update an unique field of target user
//	@Description	Update one unique data of target user: email, phone, username or document. Email and phone updates require verification.
//	@Router			/user/{id}/{field} [post]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string								false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			id				path		string								true	"User ID"
//	@Param			field			path		string								true	"Must be: email, phone, username or document"
//	@Param			payload			body		usercase.UpdateUniqueFieldsInput	true	"Value is required."
//	@Success		200				{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		403				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g AuthGateway) updateUniqueFields(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("field").
		AddFromContext(common.ActorCtxKey, "actor").
		AddFromContext(common.TargetCtxKey, "target").
		AddLanguages()

	var input usercase.UpdateUniqueFieldsInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.UpdateUniqueFields{
		AuthRepo:    queries,
		MailService: g.MailService,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// CancelPendingField godoc
//
//	@Summary		Cancel pending unique field update
//	@Description	Cancel a pending email or phone update for the target user
//	@Router			/user/me/{field}/verification [delete]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			field	path	string	true	"Field must be 'email' or 'phone'."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) cancelMyPendingField(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("field")

	var input usercase.CancelPendingFieldInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.CancelPendingField{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// GetUserID godoc
//
//	@Summary		Get user id
//	@Description	Retrieve user id by its entry. Return nil if entry does not exist.
//	@Router			/user/id [get]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			x-entry	header		string	true	"Email, username, phone or document."
//	@Success		200		{object}	presenter.Success[string]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g AuthGateway) getUserID(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddHeader("X-Entry", "entry")

	var input usercase.GetUserIDInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.GetUserID{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// GetUserEmail godoc
//
//	@Summary		Get target user email
//	@Description	Retrieve email of the target user. Require proper permission.
//	@Router			/user/{id}/email [get]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	presenter.Success[string]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) getUserEmail(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(common.TargetCtxKey, "target")

	var input usercase.GetUserEmailInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.GetUserEmail{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
