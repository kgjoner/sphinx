package authrepo

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/lib/pq"
)

func (q DAO) GetClient(clientID uuid.UUID) (*auth.Client, error) {
	raw, exists := rawQueries["GetClient"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, clientID)

	var client auth.Client
	err := row.Scan(
		&client.ID,
		&client.AudienceID,
		&client.Secret,
		&client.Name,
		pq.Array(&client.AllowedRedirectUris),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &client, nil
}