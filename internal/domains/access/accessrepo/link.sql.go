package accessrepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/datatransform"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/lib/pq"
)

func (q DAO) InsertLink(link *access.Link) error {
	raw, exists := rawQueries["CreateLink"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		link.ID,
		link.UserID,
		link.Application.ID,
		pq.Array(datatransform.ToStringArray(link.Roles)),
		link.HasConsent,
	)
	return err
}

func (q DAO) UpdateLink(link access.Link) error {
	raw, exists := rawQueries["UpdateLink"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		link.ID,
		pq.Array(datatransform.ToStringArray(link.Roles)),
		link.HasConsent,
		link.UpdatedAt,
	)
	return err
}

func (q DAO) GetUserLink(userID uuid.UUID, appID uuid.UUID) (*access.Link, error) {
	raw, exists := rawQueries["GetUserLink"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, userID, appID)

	var link access.Link
	err := row.Scan(
		&link.ID,
		&link.UserID,
		dbhandler.FromJSON(&link.Application),
		dbhandler.EnumArray(&link.Roles),
		&link.HasConsent,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &link, nil
}
