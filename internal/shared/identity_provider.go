package shared

import "github.com/kgjoner/cornucopia/v2/helpers/htypes"

type IdentityProvider interface {
	Authenticate(IdentityProviderInput) (*ExternalSubject, error)
}

type IdentityProviderInput struct {
	ProviderName string `json:"-"`
	Params       map[string]string
	Body         map[string]string

	// It should contain the full authorization header provided by the client to
	// authenticate with the identity provider, along with "Bearer " or "Basic "
	// prefix if applicable.
	Authorization string
}

/* ==============================================================================
	External Subject
============================================================================== */

// ExternalSubject represents a subject coming from an external identity provider.
type ExternalSubject struct {
	ID    string
	Kind  SubjectKind
	Email htypes.Email
	// Should be a text easy for user identification (e.g., email, username).
	Alias        string
	ProviderName string
}
