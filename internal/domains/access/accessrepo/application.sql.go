package accessrepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/datatransform"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/lib/pq"
)

func (q DAO) InsertApplication(app *access.Application) error {
	raw, exists := rawQueries["CreateApplication"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		app.ID,
		app.Name,
		pq.Array(datatransform.ToStringArray(app.PossibleRoles)),
		app.Secret,
		pq.Array(app.AllowedRedirectUris),
	)
	return err
}

func (q DAO) UpdateApplication(app access.Application) error {
	raw, exists := rawQueries["UpdateApplication"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		app.ID,
		app.Name,
		pq.Array(datatransform.ToStringArray(app.PossibleRoles)),
		pq.Array(app.AllowedRedirectUris),
	)
	return err
}

func (q DAO) GetApplicationByID(id uuid.UUID) (*access.Application, error) {
	raw, exists := rawQueries["GetApplicationByID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, id)
	var item access.Application
	err := row.Scan(
		&item.ID,
		&item.Name,
		dbhandler.EnumArray(&item.PossibleRoles),
		&item.Secret,
		pq.Array(&item.AllowedRedirectUris),
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &item, nil
}
