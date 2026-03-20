package auth

import (
	"github.com/kgjoner/cornucopia/v3/validator"
	"github.com/kgjoner/sphinx/internal/shared"
)

/* ==============================================================================
	Consent Proof
============================================================================== */

type ConsentProof struct {
	verified bool
	subID    string
	audID    string
}

// VerifyConsent checks if the principal has given consent for the client,
// and returns a ConsentProof if valid.
func VerifyConsent(p Principal, client Client) (*ConsentProof, error) {
	if !p.HasConsent || p.AudienceID != client.AudienceID {
		return nil, ErrNoConsent
	}

	return &ConsentProof{
		verified: true,
		subID:    p.ID.String(),
		audID:    client.AudienceID.String(),
	}, nil
}

// ValidFor checks if the proof is valid for the given Actor or Subject.
func (p ConsentProof) ValidFor(target any) bool {
	switch typedTarget := target.(type) {
	case Subject:
		return p.verified &&
			p.subID == typedTarget.ID.String() &&
			p.audID == typedTarget.AudienceID.String()
	case shared.Actor:
		return p.verified &&
			p.subID == typedTarget.ID.String()
		// Actor audID is from root application; they don't have a field
		// to match against audID in the proof. The actor ID is enough once
		// their permission should be already checked by policies.
	default:
		return false
	}
}

/* ==============================================================================
	Grant Proof
============================================================================== */

type GrantProof struct {
	verified bool
	subID    string
	audID    string
}

// VerifyGrant checks if the provided grant and credentials are valid,
// and returns a GrantProof if they are.
func VerifyGrant(grant Grant, client Client, credentials GrantCredentials, solver any) (*GrantProof, error) {
	defer func() { grant.IsUsed = true }()

	err := validator.Validate(credentials)
	if err != nil {
		return nil, err
	}

	// Check basic grant validity
	if !grant.isActive() {
		return nil, ErrExpiredGrant
	}

	// Check client matches
	if grant.ClientID != credentials.ClientID {
		return nil, ErrInvalidClient
	}

	// Check simple credentials matches
	if credentials.GrantType != grant.Type || credentials.Code != grant.Code || credentials.RedirectUri != grant.RedirectUri {
		return nil, shared.ErrInvalidCredentials
	}

	// Validate according to grant and client type
	if grant.Type != "authorization_code" {
		return nil, ErrUnsupportedGrantType
	}

	hasValidCredentials := false
	if credentials.IsConfidentialClient() {
		hasher, ok := solver.(shared.PasswordHasher)
		if !ok {
			return nil, ErrInvalidSolver
		}

		if hasher.DoesPasswordMatch(client.Secret.String(), credentials.AppSecret) {
			hasValidCredentials = true
		}
	}

	if credentials.IsPKCE() {
		challenger, ok := solver.(CodeChallenger)
		if !ok {
			return nil, ErrInvalidSolver
		}

		if challenger.DoesChallengeMatch(string(grant.CodeChallengeMethod), grant.CodeChallenge, credentials.CodeVerifier) {
			hasValidCredentials = true
		}
	}

	if !hasValidCredentials {
		return nil, shared.ErrInvalidCredentials
	}

	return &GrantProof{
		verified: true,
		subID:    grant.SubID.String(),
		audID:    grant.AudID.String(),
	}, nil
}

// ValidFor checks if the proof is valid for the given Grant or Principal.
func (p GrantProof) ValidFor(target any) bool {
	switch typedTarget := target.(type) {
	case Grant:
		return p.verified &&
			p.subID == typedTarget.SubID.String() &&
			p.audID == typedTarget.AudID.String()
	case Principal:
		return p.verified &&
			p.subID == typedTarget.ID.String() &&
			p.audID == typedTarget.AudienceID.String()
	default:
		return false
	}
}

/* ==============================================================================
	External Auth Proof
============================================================================== */

type ExternalLoginProof struct {
	verified      bool
	providerName  string
	providerSubID string
}

// VerifyExternalLogin checks if the principal has given consent for the client,
// and returns an ExternalLoginProof if valid.
func VerifyExternalLogin(p Principal, extSub shared.ExternalSubject) (*ExternalLoginProof, error) {
	if !p.HasConsent {
		return nil, ErrNoConsent
	}

	var exists bool
	for _, cred := range p.ExternalCredentials {
		if cred.ProviderName == extSub.ProviderName && cred.ProviderSubjectID == extSub.ID {
			exists = true
			break
		}
	}

	if !exists {
		return nil, ErrNoUserForThisSubject
	}

	return &ExternalLoginProof{
		verified:      true,
		providerName:  extSub.ProviderName,
		providerSubID: extSub.ID,
	}, nil
}

// ValidFor checks if the proof is valid for the given Principal or ExternalSubject.
func (p ExternalLoginProof) ValidFor(target any) bool {
	switch typedTarget := target.(type) {
	case Principal:
		var exists bool
		for _, cred := range typedTarget.ExternalCredentials {
			if cred.ProviderName == p.providerName && cred.ProviderSubjectID == p.providerSubID {
				exists = true
				break
			}
		}
		return p.verified && exists
	case shared.ExternalSubject:
		return p.verified &&
			p.providerSubID == typedTarget.ID &&
			p.providerName == typedTarget.ProviderName
	default:
		return false
	}
}
