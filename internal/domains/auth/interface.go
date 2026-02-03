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

	// Signing key management
	InsertSigningKey(*SigningKey) error
	UpdateSigningKey(SigningKey) error
	AcquireSigningKeyLock() (release func(), err error)
	ListActiveSigningKeys() ([]*SigningKey, error)
	GetSigningKeyByKID(string) (*SigningKey, error)
	DeactivateExpiredKeys() error
}

type TokenProvider interface {
	Generate(sub Subject) (*Tokens, error)
	Validate(token string) (*Subject, Intent, error)
}

type CodeChallenger interface {
	DoesChallengeMatch(method string, challenge string, verifier string) bool
}

type Encryptor interface {
	Encrypt(plainText []byte) ([]byte, error)
	Decrypt(cipherText []byte) ([]byte, error)
}

type KeyProvisioner interface {
	GeneratePair() (privateKeyPEM []byte, publicKeyPEM []byte, err error)
	DecodePrivate(privateKeyPEM []byte) (privateKey interface{}, err error)
	DecodePublic(publicKeyPEM []byte) (publicKey interface{}, err error)
}
