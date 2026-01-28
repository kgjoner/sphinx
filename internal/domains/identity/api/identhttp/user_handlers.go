package identhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	usercase "github.com/kgjoner/sphinx/internal/domains/identity/cases/user"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

func (g Gateway) userHandler(r chi.Router) {
	r.Post("/", g.createUser)
	r.Get("/existence", g.checkEntryExistence)
	r.Post("/request-password", g.requestPasswordReset)
	r.Post("/{id}/password", g.resetPassword)
	r.Post("/{id}/{field}/verification", g.verifyUser)

	r.With(g.mid.Authenticate, g.mid.TargetUser).Get("/me", g.getMe)
	r.With(g.mid.Authenticate, g.mid.TargetUser).Patch("/me", g.updateMyExtraData)
	r.With(g.mid.Authenticate, g.mid.TargetUser).Post("/me/password", g.changeMyPassword)
	r.With(g.mid.Authenticate, g.mid.TargetUser).Post("/me/{field}", g.updateMyUniqueFields)
	r.With(g.mid.Authenticate, g.mid.TargetUser).Delete("/me/{field}/verification", g.cancelMyPendingField)

	r.With(g.mid.Authenticate, g.mid.TargetUser).Get("/{userID}", g.getPrivateUser)
	r.With(g.mid.Authenticate, g.mid.TargetUser).Patch("/{userID}", g.updateExtraData)
	r.With(g.mid.Authenticate, g.mid.TargetUser).Post("/{userID}/{field}", g.updateUniqueFields)
	r.With(g.mid.AuthenticateApp, g.mid.TargetUser).Get("/{userID}/email", g.getUserEmail)

	// Utility endpoint to get user ID by its entry. Passed on X-Entry header.
	r.With(g.mid.AuthenticateApp).Get("/id", g.getUserID)
}

// SignUp godoc
//
//	@Summary		Sign up a new user
//	@Description	Register a new user linked to root app, and send email validation code.
//	@Router			/user [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string					false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		identity.UserCreationFields	true	"Email and password are mandatory."
//	@Success		201				{object}	presenter.Success[identity.UserLeanView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		409				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g Gateway) createUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddLanguages()

	var input usercase.SignUpInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx Repo) (any, error) {
		i := usercase.SignUp{
			IdentityRepo: tx,
			AccessRepo:   tx,
			Mailer:       g.Services.Mailer,
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
func (g Gateway) checkEntryExistence(w http.ResponseWriter, r *http.Request) {
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
		IdentityRepo: queries,
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
//	@Success		200	{object}	presenter.Success[identity.UserView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g Gateway) getMe(w http.ResponseWriter, r *http.Request) {
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
//	@Success		200	{object}	presenter.Success[identity.UserView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g Gateway) getPrivateUser(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "targetID")

	var input usercase.GetPrivateUserInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.GetPrivateUser{
		IdentityRepo: queries,
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
func (g Gateway) verifyUser(w http.ResponseWriter, r *http.Request) {
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
		IdentityRepo: queries,
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
func (g Gateway) changeMyPassword(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "targetID").
		AddLanguages()

	var input usercase.ChangePasswordInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx Repo) (any, error) {
		i := usercase.ChangePassword{
			IdentityRepo: tx,
			AuthRepo:     tx,
			Mailer:       g.Mailer,
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
func (g Gateway) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
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
		IdentityRepo: queries,
		Mailer:       g.Mailer,
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
func (g Gateway) resetPassword(w http.ResponseWriter, r *http.Request) {
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

	output, err := g.BasePool.WithTransaction(r.Context(), nil, func(tx Repo) (any, error) {
		i := usercase.ResetPassword{
			IdentityRepo: tx,
			AuthRepo:     tx,
			Mailer:       g.Services.Mailer,
		}

		return i.Execute(input)
	})
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
//	@Param			payload	body		identity.ExtraData	true	"At least one data must be defined.""
//	@Success		200		{object}	presenter.Success[identity.UserView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		422		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g Gateway) updateMyExtraData(w http.ResponseWriter, r *http.Request) {
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
//	@Param			id		path		string			true	"User ID"'
//	@Param			payload	body		identity.ExtraData	true	"At least one data must be defined.""
//	@Success		200		{object}	presenter.Success[identity.UserView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		422		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g Gateway) updateExtraData(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "targetID")

	var input usercase.UpdateExtraDataInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.UpdateExtraData{
		IdentityRepo: queries,
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
//	@Success		200				{object}	presenter.Success[identity.UserView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g Gateway) updateMyUniqueFields(w http.ResponseWriter, r *http.Request) {
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
//	@Param			payload			body		usercase.UpdateUniqueFieldInput	true	"Value is required."
//	@Success		200				{object}	presenter.Success[identity.UserView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		403				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g Gateway) updateUniqueFields(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("field").
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "targetID").
		AddLanguages()

	var input usercase.UpdateUniqueFieldInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.UpdateUniqueField{
		IdentityRepo: queries,
		Mailer:       g.Mailer,
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
func (g Gateway) cancelMyPendingField(w http.ResponseWriter, r *http.Request) {
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
		IdentityRepo: queries,
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
func (g Gateway) getUserID(w http.ResponseWriter, r *http.Request) {
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
		IdentityRepo: queries,
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
func (g Gateway) getUserEmail(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(sharedhttp.TargetIDCtxKey, "targetID")

	var input usercase.GetUserEmailInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := usercase.GetUserEmail{
		IdentityRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
