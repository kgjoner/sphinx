package baserepo

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

func (q Queries) UpsertLinks(links ...auth.Link) error {
	if len(links) == 0 {
		return nil
	}

	type formattedLink struct {
		Id            uuid.UUID   `json:"id" validate:"required"`
		AccountId     int         `json:"account_id" validate:"required"`
		ApplicationId int         `json:"application_id"`
		Roles         []auth.Role `json:"roles"`
		HasConsent    bool        `json:"has_consent"`

		CreatedAt time.Time `json:"created_at" validate:"required"`
		UpdatedAt time.Time `json:"updated_at" validate:"required"`
	}

	formattedLinks := []formattedLink{}
	for _, l := range links {
		isDuplicated := false
		for _, f := range formattedLinks {
			if f.Id == l.Id {
				isDuplicated = true
				break
			}
		}
		if isDuplicated {
			continue
		}

		formattedLink := formattedLink{
			l.Id,
			l.AccountId,
			l.Application.InternalId,
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
