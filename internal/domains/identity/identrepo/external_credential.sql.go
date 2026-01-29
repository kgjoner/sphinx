package identrepo

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/identity"
)

func (q DAO) InsertExternalCredential(credential *identity.ExternalCredential) error {
	raw, exists := rawQueries["CreateExternalCredential"]
	if !exists {
		return ErrNoQuery
	}

	q.dbtx.QueryRowContext(q.ctx, raw,
		credential.UserID,
		credential.ProviderName,
		credential.ProviderSubjectID,
		credential.LastUsedAt,
	)
	return nil
}

func (q DAO) UpdateExternalCredential(credential identity.ExternalCredential) error {
	raw, exists := rawQueries["UpdateExternalCredential"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		credential.ProviderName,
		credential.ProviderSubjectID,
		credential.LastUsedAt,
	)
	return err
}

func (q DAO) RemoveExternalCredential(userID uuid.UUID, providerName string, providerSubjectID string) error {
	raw, exists := rawQueries["RemoveExternalCredential"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		userID,
		providerName,
		providerSubjectID,
	)
	return err
}
