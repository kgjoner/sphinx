package accessrepo

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/datatransform"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/access"
)

func (q DAO) UpsertLinks(links ...access.Link) error {
	if len(links) == 0 {
		return nil
	}

	type formattedLink struct {
		ID            uuid.UUID     `json:"id" validate:"required"`
		UserID        uuid.UUID     `json:"user_id" validate:"required"`
		ApplicationID uuid.UUID     `json:"application_id"`
		Roles         []access.Role `json:"roles"`
		HasConsent    bool          `json:"has_consent"`

		CreatedAt time.Time `json:"created_at" validate:"required"`
		UpdatedAt time.Time `json:"updated_at" validate:"required"`
	}

	formattedLinks := []formattedLink{}
	for _, l := range links {
		isDuplicated := false
		for _, f := range formattedLinks {
			if f.ID == l.ID {
				isDuplicated = true
				break
			}
		}
		if isDuplicated {
			continue
		}

		formattedLink := formattedLink{
			l.ID,
			l.UserID,
			l.Application.ID,
			l.Roles,
			l.HasConsent,
			l.CreatedAt,
			l.UpdatedAt,
		}
		formattedLinks = append(formattedLinks, formattedLink)
	}

	raw, exists := rawQueries["UpsertLinks"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		datatransform.ToRawMessage(formattedLinks),
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
	var roles []access.Role
	err := row.Scan(
		&link.ID,
		&link.UserID,
		dbhandler.FromJSON(&link.Application),
		dbhandler.EnumArray(&link.Roles),
		&link.HasConsent,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	link.Roles = roles

	return &link, nil
}
