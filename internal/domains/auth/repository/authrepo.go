package authrepo

import (
	"context"

	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	psqlrepo "github.com/kgjoner/sphinx/postgres"
)

type AuthRepo struct {
	q   *psqlrepo.Queries
	ctx context.Context
}

func New(q *psqlrepo.Queries) *AuthRepo {
	return &AuthRepo{
		q: q,
	}
}

func (r *AuthRepo) AddContext(ctx context.Context) {
	r.ctx = ctx
}

/* =========================================================================
	LINK and SESSION
========================================================================= */

func (r AuthRepo) UpsertLinks(links ...auth.Link) error {
	if len(links) == 0 {
		return nil
	}

	type formattedLink struct{
		auth.Link
		ApplicationId int `json:"application_id"`
	}
	
	formattedLinks := []formattedLink{}
	for _, l := range links {
		formattedLink := formattedLink{
			l,
			l.Application.InternalId,
		}
		formattedLinks = append(formattedLinks, formattedLink)
	}

	return r.q.UpsertLinks(r.ctx, datatransform.ToRawMessage(formattedLinks))
}

func (r AuthRepo) UpsertSessions(sessions ...auth.Session) error {
	if len(sessions) == 0 {
		return nil
	}

	type formattedSession struct{
		auth.Session
		ApplicationId int `json:"application_id"`
	}
	
	formattedSessions := []formattedSession{}
	for _, s := range sessions {
		formattedSession := formattedSession{
			s,
			s.Application.InternalId,
		}
		formattedSessions = append(formattedSessions, formattedSession)
	}

	return r.q.UpsertSessions(r.ctx, datatransform.ToRawMessage(formattedSessions))
}
