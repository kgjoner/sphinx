package accesshttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/access/accesscase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

func (g gateway) linkHandler(r chi.Router) {
	authedUserR := r.With(g.Authenticate, g.TargetUser)
	authedUserR.Get("/", g.getLink)

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
//	@Success		200		{object}	httpserver.Success[access.LinkView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		404		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) getLink(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "userID").
		ParseURLParam("appID", "applicationID")

	var input accesscase.GetLinkInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.GetLink{
		AccessRepo: repo,
	}
	out, err := i.Execute(input)

	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(out, w, r, http.StatusOK)
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
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "userID").
		ParseURLParam("appID", "applicationID").
		ParseURLParam("role", "role")

	var input accesscase.AddRoleInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.AddRole{
		AccessRepo: repo,
	}
	_, err = i.Execute(input)

	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(nil, w, r, http.StatusNoContent)
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
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "userID").
		ParseURLParam("appID", "applicationID").
		ParseURLParam("role", "role")

	var input accesscase.RemoveRoleInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.RemoveRole{
		AccessRepo: repo,
	}
	_, err = i.Execute(input)

	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(nil, w, r, http.StatusNoContent)
}
