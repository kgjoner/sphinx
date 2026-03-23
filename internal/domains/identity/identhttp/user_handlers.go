package identhttp

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	_ "github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/sphinx/internal/domains/identity/identcase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

func (g gateway) userHandler(r chi.Router) {
	r.Post("/", g.createUser)
	r.Get("/existence", g.checkEntryExistence)
	r.Post("/request-password", g.requestPasswordReset)
	r.Post("/{id}/password", g.resetPassword)
	r.Post("/{id}/{field}/verification", g.verifyUser)

	r.With(g.Authenticate).Get("/", g.listUsers)

	r.With(g.Authenticate, g.TargetUser).Get("/me", g.getMe)
	r.With(g.Authenticate, g.TargetUser).Patch("/me", g.updateMyExtraData)
	r.With(g.Authenticate, g.TargetUser).Post("/me/password", g.changeMyPassword)
	r.With(g.Authenticate, g.TargetUser).Post("/me/{field}", g.updateMyUniqueField)
	r.With(g.Authenticate, g.TargetUser).Delete("/me/{field}/verification", g.cancelMyPendingField)

	r.With(g.AuthenticateAny, g.TargetUser).Get("/{userID}", g.getPrivateUser)
	r.With(g.Authenticate, g.TargetUser).Patch("/{userID}", g.updateExtraData)
	r.With(g.Authenticate, g.TargetUser).Post("/{userID}/{field}", g.updateUniqueField)
	r.With(g.AuthenticateApp, g.TargetUser).Get("/{userID}/email", g.getUserEmail)

	// Utility endpoint to get user ID by its entry. Passed on X-Entry header.
	r.With(g.AuthenticateApp).Get("/id", g.getUserID)
}

// SignUp godoc
//
//	@Summary		Sign up a new user
//	@Description	Register a new user linked to root app, and send email validation code.
//	@Router			/user [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header		string						false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			payload			body		identity.UserCreationFields	true	"Email and password are mandatory."
//	@Success		201				{object}	httpserver.SuccessResponse[identity.UserLeanView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		409				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g gateway) createUser(w http.ResponseWriter, r *http.Request) {
	var input identcase.SignUpInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		Languages(&input.Languages)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	output, err := g.PGPool.WithTx(r.Context(), nil, func(tx *sql.Tx) (any, error) {
		identRepo := g.IdentFactory.NewDAO(r.Context(), tx)
		accessRepo := g.AccessFactory.NewDAO(r.Context(), tx)

		i := identcase.SignUp{
			IdentityRepo: identRepo,
			AccessRepo:   accessRepo,
			PwHasher:     g.PwHasher,
			Mailer:       g.Mailer,
		}

		return i.Execute(input)
	})

	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusCreated)
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
//	@Success		200		{object}	httpserver.SuccessResponse[bool]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) checkEntryExistence(w http.ResponseWriter, r *http.Request) {
	var input identcase.CheckEntryExistenceInput
	c := httpserver.Bind(r).
		Header("X-Entry", &input.Entry)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.CheckEntryExistence{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
//	@Success		200	{object}	httpserver.SuccessResponse[identity.UserView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) getMe(w http.ResponseWriter, r *http.Request) {
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
//	@Success		200	{object}	httpserver.SuccessResponse[identity.UserView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) getPrivateUser(w http.ResponseWriter, r *http.Request) {
	var input identcase.GetPrivateUserInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.TargetID)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.GetPrivateUser{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
//	@Param			payload	body	identcase.VerifyUserInput	true	"Code is required."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) verifyUser(w http.ResponseWriter, r *http.Request) {
	var input identcase.VerifyUserInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		PathParam("id", &input.UserID).
		PathParam("field", &input.VerificationKind)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.VerifyUser{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusNoContent)
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
//	@Param			payload			body	identcase.ChangePasswordInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) changeMyPassword(w http.ResponseWriter, r *http.Request) {
	var input identcase.ChangePasswordInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.TargetID).
		Languages(&input.Languages)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	output, err := g.PGPool.WithTx(r.Context(), nil, func(tx *sql.Tx) (any, error) {
		i := identcase.ChangePassword{
			IdentityRepo: g.IdentFactory.NewDAO(r.Context(), tx),
			AuthRepo:     g.AuthFactory.NewDAO(r.Context(), tx),
			PwHasher:     g.PwHasher,
			Mailer:       g.Mailer,
		}

		return i.Execute(input)
	})
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusNoContent)
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
//	@Param			payload			body	identcase.RequestPasswordResetInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input identcase.RequestPasswordResetInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		Languages(&input.Languages)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.RequestPasswordReset{
		IdentityRepo: repo,
		Mailer:       g.Mailer,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusNoContent)
}

// ResetPassword godoc
//
//	@Summary		Update password with a reset code.
//	@Description	Update password with a reset code.
//	@Router			/user/{id}/password [post]
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			accept-language	header	string							false	"Used to define mailing language. Example: pt-br, pt;q=0.9, en;q=0.5"
//	@Param			id				path	string							true	"User ID"
//	@Param			payload			body	identcase.ResetPasswordInput	true	"Old password and new one."
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) resetPassword(w http.ResponseWriter, r *http.Request) {
	var input identcase.ResetPasswordInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		PathParam("id", &input.UserID).
		Languages(&input.Languages)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	output, err := g.PGPool.WithTx(r.Context(), nil, func(tx *sql.Tx) (any, error) {
		i := identcase.ResetPassword{
			IdentityRepo: g.IdentFactory.NewDAO(r.Context(), tx),
			AuthRepo:     g.AuthFactory.NewDAO(r.Context(), tx),
			PwHasher:     g.PwHasher,
			Mailer:       g.Mailer,
		}

		return i.Execute(input)
	})
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusNoContent)
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
//	@Success		200		{object}	httpserver.SuccessResponse[identity.UserView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		422		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) updateMyExtraData(w http.ResponseWriter, r *http.Request) {
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
//	@Param			id		path		string				true	"User ID"'
//	@Param			payload	body		identity.ExtraData	true	"At least one data must be defined.""
//	@Success		200		{object}	httpserver.SuccessResponse[identity.UserView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		422		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) updateExtraData(w http.ResponseWriter, r *http.Request) {
	var input identcase.UpdateExtraDataInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.TargetID)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.UpdateExtraData{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
}

// UpdateMyUniqueField godoc
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
//	@Param			payload			body		identcase.UpdateUniqueFieldInput	true	"Value is required."
//	@Success		200				{object}	httpserver.SuccessResponse[identity.UserView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g gateway) updateMyUniqueField(w http.ResponseWriter, r *http.Request) {
	g.updateUniqueField(w, r)
}

// UpdateUniqueField godoc
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
//	@Param			payload			body		identcase.UpdateUniqueFieldInput	true	"Value is required."
//	@Success		200				{object}	httpserver.SuccessResponse[identity.UserView]
//	@Failure		400				{object}	apperr.AppError
//	@Failure		401				{object}	apperr.AppError
//	@Failure		403				{object}	apperr.AppError
//	@Failure		422				{object}	apperr.AppError
//	@Failure		500				{object}	apperr.AppError
func (g gateway) updateUniqueField(w http.ResponseWriter, r *http.Request) {
	var input identcase.UpdateUniqueFieldInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		PathParam("field", &input.Field).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.TargetID).
		Languages(&input.Languages)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.UpdateUniqueField{
		IdentityRepo: repo,
		Mailer:       g.Mailer,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
func (g gateway) cancelMyPendingField(w http.ResponseWriter, r *http.Request) {
	var input identcase.CancelPendingFieldInput
	c := httpserver.Bind(r).
		PathParam("field", &input.Field)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.CancelPendingField{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusNoContent)
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
//	@Success		200		{object}	httpserver.SuccessResponse[string]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) getUserID(w http.ResponseWriter, r *http.Request) {
	var input identcase.GetUserIDInput
	c := httpserver.Bind(r).
		Header("X-Entry", &input.Entry)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.GetUserID{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
//	@Success		200	{object}	httpserver.SuccessResponse[string]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) getUserEmail(w http.ResponseWriter, r *http.Request) {
	var input identcase.GetUserEmailInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.TargetIDCtxKey, &input.TargetID)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.GetUserEmail{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
}

// ListUsers godoc
//
//	@Summary		List users
//	@Description	Retrieve a paginated list of users, from most recent to oldest. Optionally filter by search term.
//	@Router			/user [get]
//	@Tags			User
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			s		query		string	false	"Search filter (applied to username, email, name, or surname)"
//	@Param			view	query		string	false	"View type (lean or full). Default is lean."
//	@Param			limit	query		int		false	"Number of results per page (default: 20)"
//	@Param			offset	query		int		false	"Number of results to skip (default: 0)"
//	@Success		200		{object}	httpserver.SuccessResponse[prim.PaginatedData[identity.UserLeanView]]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) listUsers(w http.ResponseWriter, r *http.Request) {
	var input identcase.ListUsersInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		QueryParam("s", &input.SearchFilter).
		QueryParam("view", &input.View).
		Pagination(&input.Pagination)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.IdentFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := identcase.ListUsers{
		IdentityRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
}
