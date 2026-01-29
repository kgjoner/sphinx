package accesshttp

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/access/accesscase"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

type gateway struct {
	Dependencies
}

type Dependencies struct {
	AccessPool shared.RepoPool[access.Repo]
	*sharedhttp.Middleware
}

func Raise(router chi.Router, deps Dependencies) {
	gtw := &gateway{
		deps,
	}
	router.Route("/application", gtw.applicationHandler)
	router.Route("/user/{userID}/link/{appID}", gtw.linkHandler)

	// Legacy route for backwards compatibility
	router.With(gtw.AuthenticateApp, gtw.TargetUser).Patch("/user/{userID}/permission", gtw.editUserPermissions)
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
//	@Param			payload	body	accesscase.EditUserPermissionsInput	true	"At least one of roles and grantings must be defined"
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		403	{object}	apperr.AppError
//	@Failure		422	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) editUserPermissions(w http.ResponseWriter, r *http.Request) {
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

	repo := g.AccessPool.NewDAO(r.Context())
	var output bool
	if input.ShouldRemove.Bool {
		i := accesscase.RemoveRole{
			AccessRepo: repo,
		}
		output, err = i.Execute(accesscase.RemoveRoleInput{
			UserID:        input.UserID,
			ApplicationID: input.Actor.ID, // the actor is the application
			Role:          input.Roles[0],
			Actor:         input.Actor,
		})
	} else {
		i := accesscase.AddRole{
			AccessRepo: repo,
		}
		output, err = i.Execute(accesscase.AddRoleInput{
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
