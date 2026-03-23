package accesshttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/access/accesscase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

func (g gateway) linkHandler(r chi.Router) {
	authedAnyR := r.With(g.AuthenticateAny, g.TargetUser)
	authedAnyR.Get("/", g.getLink)

	authedAppR := r.With(g.AuthenticateApp, g.TargetUser)
	authedAppR.Put("/role/{role}", g.addRole)
	authedAppR.Delete("/role/{role}", g.removeRole)
}

// GetLink godoc
//
//	@Summary		Get a user link
//	@Description	Retrieve a link between a user and an application. Must be used by authenticated applications with proper permissions.
//	@Router			/user/{userID}/link/{appID} [get]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Param			appID	path		string	true	"Application ID"
//	@Success		200		{object}	httpserver.SuccessResponse[access.LinkView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		404		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) getLink(w http.ResponseWriter, r *http.Request) {
	var input accesscase.GetLinkInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.UserID).
		PathParam("appID", &input.ApplicationID)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.GetLink{
		AccessRepo: repo,
	}
	out, err := i.Execute(input)

	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(out, w, r, http.StatusOK)
}

// AddRole godoc
//
//	@Summary		Add roles to a user link
//	@Description	Add roles to a user link. Must be used by authenticated applications with proper permissions.
//	@Router			/user/{userID}/link/{appID}/role/{role} [put]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	string	true	"User ID"
//	@Param			appID	path	string	true	"Application ID"
//	@Param			role	path	string	true	"Role to add"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) addRole(w http.ResponseWriter, r *http.Request) {
	var input accesscase.AddRoleInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.UserID).
		PathParam("appID", &input.ApplicationID).
		PathParam("role", &input.Role)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.AddRole{
		AccessRepo: repo,
	}
	_, err := i.Execute(input)

	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(nil, w, r, http.StatusNoContent)
}

// RemoveRole godoc
//
//	@Summary		Remove roles from a user link
//	@Description	Remove roles from a user link. Must be used by authenticated applications with proper permissions.
//	@Router			/user/{userID}/link/{appID}/role/{role} [delete]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	string	true	"User ID"
//	@Param			appID	path	string	true	"Application ID"
//	@Param			role	path	string	true	"Role to remove"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) removeRole(w http.ResponseWriter, r *http.Request) {
	var input accesscase.RemoveRoleInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TargetIDCtxKey, &input.UserID).
		PathParam("appID", &input.ApplicationID).
		PathParam("role", &input.Role)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.RemoveRole{
		AccessRepo: repo,
	}
	_, err := i.Execute(input)

	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(nil, w, r, http.StatusNoContent)
}
