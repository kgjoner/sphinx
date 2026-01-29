package authrepo

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/lib/pq"
)

func (q DAO) GetPrincipal(subID uuid.UUID, audID uuid.UUID) (*auth.Principal, error) {
	raw, exists := rawQueries["GetPrincipal"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, subID, audID)

	var session auth.Principal
	err := row.Scan(
		&session.ID,
		&session.Kind,
		&session.Password,
		&session.Email,
		&session.Name,
		&session.AudienceID,
		pq.Array(&session.Roles),
		&session.HasConsent,
		&session.IsActive,
		dbhandler.FromJSON(&session.ExternalCredentials),
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (q DAO) GetPrincipalByEntry(entry shared.Entry, audID uuid.UUID) (*auth.Principal, error) {
	raw, exists := rawQueries["GetPrincipalByEntry"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, entry, audID)

	var session auth.Principal
	err := row.Scan(
		&session.ID,
		&session.Kind,
		&session.Password,
		&session.Email,
		&session.Name,
		&session.AudienceID,
		pq.Array(&session.Roles),
		&session.HasConsent,
		&session.IsActive,
		dbhandler.FromJSON(&session.ExternalCredentials),
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (q DAO) GetPrincipalByExtSubID(providerName string, extSubID string, audID uuid.UUID) (*auth.Principal, error) {
	raw, exists := rawQueries["GetPrincipalByExtSubID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, providerName, extSubID, audID)

	var session auth.Principal
	err := row.Scan(
		&session.ID,
		&session.Kind,
		&session.Password,
		&session.Email,
		&session.Name,
		&session.AudienceID,
		pq.Array(&session.Roles),
		&session.HasConsent,
		&session.IsActive,
		dbhandler.FromJSON(&session.ExternalCredentials),
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
