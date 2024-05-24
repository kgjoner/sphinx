package authrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	psqlrepo "github.com/kgjoner/sphinx/postgres"
)

type AuthFactory struct {
	q *psqlrepo.Queries
}

func NewFactory(q *psqlrepo.Queries) *AuthFactory {
	return &AuthFactory{
		q: q,
	}
}

func (f AuthFactory) New(ctx context.Context) authcase.AuthRepo {
	return &AuthRepo{
		q:   f.q,
		ctx: ctx,
	}
}

type AuthRepo struct {
	q   *psqlrepo.Queries
	ctx context.Context
}

/* =========================================================================
	LINK and SESSION
========================================================================= */

func (r AuthRepo) UpsertLinks(links ...auth.Link) error {
	if len(links) == 0 {
		return nil
	}

	type formattedLink struct {
		Id            uuid.UUID   `json:"id" validate:"required"`
		AccountId     int         `json:"account_id" validate:"required"`
		ApplicationId int         `json:"application_id"`
		Roles         []auth.Role `json:"roles"`
		Grantings     []string    `json:"grantings"`

		OAuthCode      string          `json:"oauth_code"`
		OAuthExpiresAt htypes.NullTime `json:"oauth_expires_at"`

		CreatedAt time.Time `json:"created_at" validate:"required"`
		UpdatedAt time.Time `json:"updated_at" validate:"required"`
	}

	formattedLinks := []formattedLink{}
	for _, l := range links {
		formattedLink := formattedLink{
			l.Id,
			l.AccountId,
			l.Application.InternalId,
			l.Roles,
			l.Grantings,
			l.OAuthCode,
			l.OAuthExpiresAt,
			l.CreatedAt,
			l.UpdatedAt,
		}
		formattedLinks = append(formattedLinks, formattedLink)
	}

	return r.q.UpsertLinks(r.ctx, datatransform.ToRawMessage(formattedLinks))
}

func (r AuthRepo) UpsertSessions(sessions ...auth.Session) error {
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

	return r.q.UpsertSessions(r.ctx, datatransform.ToRawMessage(formattedSessions))
}
