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
	return r.q.UpsertLinks(r.ctx, datatransform.ToRawMessage(links))
}

func (r AuthRepo) UpsertSessions(sessions ...auth.Session) error {
	if len(sessions) == 0 {
		return nil
	}
	return r.q.UpsertSessions(r.ctx, datatransform.ToRawMessage(sessions))
}
