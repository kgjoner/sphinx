package authrepo

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	psqlrepo "github.com/kgjoner/sphinx/postgres"
)

func (r AuthRepo) InsertApplication(app auth.Application) (int, error) {
	return r.q.CreateApplication(r.ctx, psqlrepo.CreateApplicationParams{
		ID:                  app.Id,
		Name:                app.Name,
		Grantings:           app.Grantings,
		Secret:              app.Secret,
		AllowedRedirectUris: app.AllowedRedirectUris,
	})
}

func (r AuthRepo) UpdateApplication(app auth.Application) error {
	return r.q.UpdateApplication(r.ctx, psqlrepo.UpdateApplicationParams{
		ID:                  app.Id,
		Name:                app.Name,
		Grantings:           app.Grantings,
		AllowedRedirectUris: app.AllowedRedirectUris,
	})
}

func (r AuthRepo) GetApplicationById(id uuid.UUID) (*auth.Application, error) {
	return dbhandler.HandleSingleQuery[auth.Application](
		r.q.GetApplicationById(r.ctx, id),
	)
}
