package baserepo

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

func (q Queries) UpsertSessions(sessions ...auth.Session) error {
	if len(sessions) == 0 {
		return nil
	}

	type formattedSession struct {
		Id                             uuid.UUID       `json:"id" validate:"required"`
		AccountId                      int             `json:"account_id" validate:"required"`
		ApplicationId                  int             `json:"application_id"`
		RefreshToken                   string          `json:"refresh_token" validate:"required"`
		RefreshedAt                    htypes.NullTime `json:"refreshed_at"`
		ElapsedMinutesBetweenRefreshes []int           `json:"elapsed_minutes_between_refreshes"`
		RefreshesCount                 int             `json:"refreshes_count"`
		Device                         string          `json:"device" validate:"required"`
		Ip                             string          `json:"ip"`
		IsActive                       bool            `json:"is_active"`
		TerminatedAt                   htypes.NullTime `json:"terminated_at"`
		CreatedAt                      time.Time       `json:"created_at" validate:"required"`
		UpdatedAt                      time.Time       `json:"updated_at" validate:"required"`
	}

	formattedSessions := []formattedSession{}
	for _, s := range sessions {
		isDuplicated := false
		for _, f := range formattedSessions {
			if f.Id == s.Id {
				isDuplicated = true
				break
			}
		}
		if isDuplicated {
			continue
		}

		formattedSession := formattedSession{
			s.Id,
			s.AccountId,
			s.Application.InternalId,
			s.RefreshToken,
			s.RefreshedAt,
			s.ElapsedMinutesBetweenRefreshes,
			s.RefreshesCount,
			s.Device,
			s.Ip,
			s.IsActive,
			s.TerminatedAt,
			s.CreatedAt,
			s.UpdatedAt,
		}
		formattedSessions = append(formattedSessions, formattedSession)
	}

	raw, exists := rawQueries["UpsertSessions"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.db.ExecContext(q.ctx, raw,
		datatransform.ToRawMessage(formattedSessions),
	)
	return err
}