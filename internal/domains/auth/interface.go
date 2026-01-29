package auth

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Repo interface {
	InsertSession(*Session) error
	UpdateSession(Session) error
	GetSessionByID(uuid.UUID) (*Session, error)
	TerminateAllSubjectSessions(subjectID uuid.UUID) error

	// It should return nil if subject exists but no relation with the given audience
	GetPrincipal(subID uuid.UUID, audID uuid.UUID) (*Principal, error)
	// It should return nil if subject exists but no relation with the given audience
	GetPrincipalByEntry(subEntry shared.Entry, audID uuid.UUID) (*Principal, error)
	// It should return nil if subject exists but no relation with the given audience
	GetPrincipalByExtSubID(providerName string, extSubID string, audID uuid.UUID) (*Principal, error)

	GetClient(uuid.UUID) (*Client, error)
}

type TokenProvider interface {
	Generate(sub Subject) (*Tokens, error)
	Validate(token string) (*Subject, Intent, error)
}

type CodeChallenger interface {
	DoesChallengeMatch(method ChallengeMethod, challenge string, verifier string) bool
}
