package baserepo

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/datatransform"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

func (q DAO) UpsertLinks(links ...auth.Link) error {
	if len(links) == 0 {
		return nil
	}

	type formattedLink struct {
		ID            uuid.UUID   `json:"id" validate:"required"`
		UserID        int         `json:"user_id" validate:"required"`
		ApplicationID int         `json:"application_id"`
		Roles         []auth.Role `json:"roles"`
		HasConsent    bool        `json:"has_consent"`

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
			l.Application.InternalID,
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

	_, err := q.db.ExecContext(q.ctx, raw,
		datatransform.ToRawMessage(formattedLinks),
	)
	return err
}
