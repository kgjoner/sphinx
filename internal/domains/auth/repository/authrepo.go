package authrepo

import (
	"context"

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
