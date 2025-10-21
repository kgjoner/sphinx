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
	r.Post("/password/request", g.requestPasswordReset)
	r.Post("/{id}/password", g.resetPassword)

	r.With(g.mid.Authenticate, g.mid.Target).Get("/", g.getPrivateUser)
	r.With(g.mid.Authenticate, g.mid.Target).Patch("/", g.updateExtraData)
	r.With(g.mid.Authenticate, g.mid.Target).Patch("/unique", g.updateUniqueFields)
	r.With(g.mid.Authenticate).Patch("/password", g.changePassword)

	r.With(g.mid.AuthenticateApp, g.mid.Target).Patch("/permission", g.editUserPermissions)
	r.With(g.mid.AuthenticateApp, g.mid.Target).Get("/email", g.getUserEmail)
	r.With(g.mid.AuthenticateApp).Get("/id", g.getUserID)

	r.Get("/existence", g.checkEntryExistence)
	r.Patch("/{id}/verification", g.verifyUser)
	r.Delete("/{id}/pending/{field}", g.cancelPendingField)
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
//	@Success		200				{object}	presenter.Success[auth.User]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
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

// GetPrivateAccounnt godoc
//
//	@Summary		Get user private data
//	@Description	Retrieve private data associated with logged user or target one, if x-target header is informed. The latter require special permission.
//	@Router			/user [get]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header		string	false	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible. If not informed, it will use the logged user."
//	@Success		200			{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400			{object}	apperr.AppError
//	@Failure		401			{object}	apperr.AppError
//	@Failure		403			{object}	apperr.AppError
//	@Failure		500			{object}	apperr.AppError
func (g AuthGateway) getPrivateUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor().
		AddTarget()

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

// VerifyAccounnt godoc
//
//	@Summary		Verify user data
//	@Description	Verify email or phone of target user
//	@Router			/user/{id}/verification [patch]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"User ID"
//	@Param			payload	body	usercase.VerifyUserInput	true	"Verification kind must be email or phone."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) verifyUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("id", "userID")

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

// ChangePassword godoc
//
//	@Summary		Update password of logged user.
//	@Description	Update password of logged user. It must provide the current password.
//	@Router			/user/password [patch]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header	string							false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body	usercase.ChangePasswordInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) changePassword(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddActor().
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
//	@Router			/user/password/request [post]
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
//	@Summary		Add or remove roles and grantings
//	@Description	Add or remove roles and/or grantings of the target user.
//	@Router			/user/permission [patch]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header	string								false	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible. If not informed, it will use the logged user."
//	@Param			payload		body	usercase.EditUserPermissionsInput	true	"At least one of roles and grantings must be defined"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) editUserPermissions(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddApplication().
		AddTarget()

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

// UpdateExtraData godoc
//
//	@Summary		Update extra data of target user
//	@Description	Update non unique data like name, surname and address of target user
//	@Router			/user [patch]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header		string			false	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible. If not informed, it will use the logged user."
//	@Param			payload		body		auth.ExtraData	true	"At least one data must be defined.""
//	@Success		200			{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400			{object}	apperr.AppError
//	@Failure		401			{object}	apperr.AppError
//	@Failure		403			{object}	apperr.AppError
//	@Failure		500			{object}	apperr.AppError
func (g AuthGateway) updateExtraData(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddTarget().
		AddActor()

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

// UpdateUniqueFields godoc
//
//	@Summary		Update unique fields of target user
//	@Description	Update unique data like email, phone, username and document of target user. Email and phone updates require verification.
//	@Router			/user/unique [patch]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			x-target		header		string					false	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible. If not informed, it will use the logged user."
//	@Param			accept-language	header		string					false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		auth.UserUniqueFields	true	"At least one field must be defined."
//	@Success		200				{object}	presenter.Success[auth.UserPrivateView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		403				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g AuthGateway) updateUniqueFields(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddTarget().
		AddActor().
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
//	@Router			/user/{id}/pending/{field} [delete]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"User ID"
//	@Param			field	path	string	true	"Field must be 'email' or 'phone'."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) cancelPendingField(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("id", "userID").
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

// GetAccounntID godoc
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
//	@Description	Retrieve email of the target user.
//	@Router			/user/email [get]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			x-target	header		string	false	"Beyond common entries (email, username, phone and document), it accepts ID as well. It is recommended use ID or username whenever possible. If not informed, it will use the logged user."
//	@Success		200			{object}	presenter.Success[string]
//	@Failure		400			{object}	apperr.AppError
//	@Failure		401			{object}	apperr.AppError
//	@Failure		403			{object}	apperr.AppError
//	@Failure		500			{object}	apperr.AppError
func (g AuthGateway) getUserEmail(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddTarget()

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
