package accesshttp

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/access"
	linkcase "github.com/kgjoner/sphinx/internal/domains/access/cases/link"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

type Gateway struct {
	AccessPool shared.RepoPool[access.Repo]
	mid        *sharedhttp.Middleware
}

func Raise(router chi.Router, pool shared.RepoPool[access.Repo], mid *sharedhttp.Middleware) {
	gtw := &Gateway{
		pool,
		mid,
	}

	router.Route("/application", gtw.applicationHandler)
	router.Route("/user/{userID}/link/{appID}", gtw.linkHandler)

	// Legacy route for backwards compatibility
	router.With(gtw.mid.AuthenticateApp, gtw.mid.TargetUser).Patch("/user/{userID}/permission", gtw.editUserPermissions)
}

// EditUserPermissions godoc
//
//	@Summary		(DEPRECATED) Add or remove roles
//	@Description	Use /link/{userID}/role/{role} endpoint instead. Keeping for legacy support only.
//	@Router			/user/{id}/permission [patch]
//	@Tags			User
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string								true	"User ID"
//	@Param			payload	body	linkcase.EditUserPermissionsInput	true	"At least one of roles and grantings must be defined"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g Gateway) editUserPermissions(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		ParseURLParam("userID")

	var input struct {
		UserID       uuid.UUID     `json:"-"`
		Actor        shared.Actor  `json:"-"`
		Roles        []access.Role `json:"roles"`
		ShouldRemove sql.NullBool  `json:"shouldRemove"`
	}
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	var output bool
	if input.ShouldRemove.Bool {
		i := linkcase.RemoveRole{
			AccessRepo: queries,
		}
		output, err = i.Execute(linkcase.RemoveRoleInput{
			UserID:        input.UserID,
			ApplicationID: input.Actor.ID, // the actor is the application
			Role:          input.Roles[0],
			Actor:         input.Actor,
		})
	} else {
		i := linkcase.AddRole{
			AccessRepo: queries,
		}
		output, err = i.Execute(linkcase.AddRoleInput{
			UserID:        input.UserID,
			ApplicationID: input.Actor.ID, // the actor is the application
			Role:          input.Roles[0],
			Actor:         input.Actor,
		})
	}

	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r, http.StatusNoContent)
}
