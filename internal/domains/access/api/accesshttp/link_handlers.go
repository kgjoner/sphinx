package accesshttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	linkcase "github.com/kgjoner/sphinx/internal/domains/access/cases/link"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

func (g Gateway) linkHandler(r chi.Router) {
	authedUserR := r.With(g.mid.Authenticate, g.mid.TargetUser)
	authedUserR.Get("/", g.getLink)

	authedAppR := r.With(g.mid.AuthenticateApp, g.mid.TargetUser)
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
//	@Param			userID		path	string								true	"User ID"
//	@Param			appID		path	string								true	"Application ID"
//	@Success		200	{object}	access.LinkView
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		404	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g Gateway) getLink(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "userID").
		ParseURLParam("appID", "applicationID")

	var input linkcase.GetLinkInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	i := linkcase.GetLink{
		AccessRepo: queries,
	}
	out, err := i.Execute(input)

	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(out, w, r, http.StatusOK)
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
//	@Param			userID		path	string								true	"User ID"
//	@Param			appID		path	string								true	"Application ID"
//	@Param			role		path	string								true	"Role to add"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g Gateway) addRole(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "userID").
		ParseURLParam("appID", "applicationID").
		ParseURLParam("role", "role")

	var input linkcase.AddRoleInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	i := linkcase.AddRole{
		AccessRepo: queries,
	}
	_, err = i.Execute(input)

	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(nil, w, r, http.StatusNoContent)
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
//	@Param			userID		path	string								true	"User ID"
//	@Param			appID		path	string								true	"Application ID"
//	@Param			role		path	string								true	"Role to remove"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g Gateway) removeRole(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TargetIDCtxKey, "userID").
		ParseURLParam("appID", "applicationID").
		ParseURLParam("role", "role")

	var input linkcase.RemoveRoleInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	i := linkcase.RemoveRole{
		AccessRepo: queries,
	}
	_, err = i.Execute(input)

	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(nil, w, r, http.StatusNoContent)
}
