package baserepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/lib/pq"
)

func (q Queries) InsertApplication(app *auth.Application) error {
	raw, exists := rawQueries["CreateApplication"]
	if !exists {
		return ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw,
		app.Id,
		app.Name,
		pq.Array(app.PossibleRoles),
		app.Secret,
		pq.Array(app.AllowedRedirectUris),
	)
	err := row.Scan(&app.InternalId)
	return err
}

func (q Queries) UpdateApplication(app auth.Application) error {
	raw, exists := rawQueries["UpdateApplication"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.db.ExecContext(q.ctx, raw,
		app.Id,
		app.Name,
		pq.Array(app.PossibleRoles),
		pq.Array(app.AllowedRedirectUris),
	)
	return err
}

func (q Queries) GetApplicationById(id uuid.UUID) (*auth.Application, error) {
	raw, exists := rawQueries["GetApplicationById"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, id)
	var item auth.Application
	err := row.Scan(
		&item.InternalId,
		&item.Id,
		&item.Name,
		pq.Array(&item.PossibleRoles),
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
