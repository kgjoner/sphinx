package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/cornucopia/v2/utils/pwdgen"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/shared"
)

/* ==============================================================================
	Principal
============================================================================== */

type Principal struct {
	ID                  uuid.UUID
	Kind                shared.SubjectKind
	Password            shared.HashedPassword
	Email               htypes.Email
	Name                string
	AudienceID          uuid.UUID
	Roles               []string
	HasConsent          bool
	IsActive            bool
	ExternalCredentials []struct {
		ProviderName      string
		ProviderSubjectID string
	}
}

/* ==============================================================================
	Client
============================================================================== */

type Client struct {
	ID                  uuid.UUID
	AudienceID          uuid.UUID
	Secret              shared.HashedPassword
	Name                string
	AllowedRedirectUris []string
}

/* ==============================================================================
	Subject
============================================================================== */

type Subject struct {
	ID         uuid.UUID
	Kind       shared.SubjectKind
	Email      htypes.Email
	Name       string
	AudienceID uuid.UUID
	Roles      []string
	SessionID  uuid.UUID
}

/* ==============================================================================
	Tokens
============================================================================== */

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

/* ==============================================================================
	Intent
============================================================================== */

type Intent string

const (
	IntentAccess  Intent = "access"
	IntentRefresh Intent = "refresh"
)

func (i Intent) Enumerate() any {
	return []Intent{
		IntentAccess,
		IntentRefresh,
	}
}

/* ==============================================================================
	Algorithm
============================================================================== */

type Algorithm string

const (
	RS256 Algorithm = "RS256"
	HS256 Algorithm = "HS256"
)

/* ==============================================================================
	Challenge Method
============================================================================== */

type ChallengeMethod string

const (
	ChallengePlain ChallengeMethod = "plain"
	ChallengeS256  ChallengeMethod = "S256"
)

/* ==============================================================================
	Grant Input
============================================================================== */

type GrantInput struct {
	Type                string          `json:"grant_type" validate:"required,oneof=authorization_code"`
	ClientID            uuid.UUID       `json:"client_id" validate:"required"`
	RedirectUri         string          `json:"redirect_uri" validate:"required,uri"`
	CodeChallenge       string          `json:"code_challenge"`
	CodeChallengeMethod ChallengeMethod `json:"code_challenge_method"`
}

/* ==============================================================================
	Grant
============================================================================== */

// Represents an OAuth 2.0 authorization grant. It should have an ephemeral duration,
// typically persisted in cache with a TTL. All fields are kept in JSON serialization
// to facilitate caching as JSON; be careful when sending them out.
type Grant struct {
	Type        string    `json:"grantType" validate:"required,oneof=authorization_code"`
	SubID       uuid.UUID `json:"subID" validate:"required"`
	AudID       uuid.UUID `json:"audID" validate:"required"`
	ClientID    uuid.UUID `json:"clientID" validate:"required"`
	ExpiresAt   time.Time `json:"expiresAt" validate:"required"`
	RedirectUri string    `json:"redirectUri" validate:"required,uri"`
	IsUsed      bool      `json:"isUsed"`

	Code          string `json:"code" validate:"required"`
	CodeChallenge string `json:"codeChallenge"`
	// Plain is not recommended, but part of spec. Default is S256.
	CodeChallengeMethod ChallengeMethod `json:"codeChallengeMethod"`
}

func NewGrant(g GrantInput, actor shared.Actor, client Client, proof *ConsentProof) (*Grant, error) {
	if !proof.ValidFor(actor) {
		return nil, ErrNoConsent
	}

	if g.ClientID != client.ID {
		return nil, ErrInvalidClient
	}

	// Validate redirect URI is allowed by application
	var hasFound bool
	for _, uri := range client.AllowedRedirectUris {
		if uri == g.RedirectUri {
			hasFound = true
			break
		}
	}
	if !hasFound {
		return nil, ErrInvalidRedirectURI
	}

	var code string
	switch g.Type {
	case "authorization_code":
		// TODO: move code generation to an extenal layer
		code = pwdgen.GeneratePassword(42, "lower", "upper", "number")
		if g.CodeChallenge != "" && g.CodeChallengeMethod == "" {
			g.CodeChallengeMethod = ChallengeS256 // Default to S256 if not specified
		}
	}

	now := time.Now()
	grant := &Grant{
		Type:                g.Type,
		Code:                code,
		SubID:               actor.ID,
		AudID:               actor.AudienceID,
		ClientID:            client.ID,
		ExpiresAt:           now.Add(time.Second * time.Duration(config.Env.AUTH_GRANT_LIFETIME_IN_SEC)),
		RedirectUri:         g.RedirectUri,
		CodeChallenge:       g.CodeChallenge,
		CodeChallengeMethod: g.CodeChallengeMethod,
	}

	return grant, validator.Validate(grant)
}

func (g Grant) IsValid() error {
	errs := make(map[string]string)

	if g.Type == "authorization_code" && g.CodeChallengeMethod != "" && g.CodeChallenge == "" {
		errs["code_challenge"] = "code_challenge is required when code_challenge_method is set"
	}

	if len(errs) != 0 {
		return apperr.NewMapError(errs)
	}

	return nil
}

func (g Grant) isExpired() bool {
	return g.ExpiresAt.Before(time.Now())
}

func (g Grant) isActive() bool {
	return !g.IsUsed && !g.isExpired()
}

/* ==============================================================================
	Grant Credentials
============================================================================== */

type GrantCredentials struct {
	GrantType   string    `json:"grant_type" validate:"required,oneof=authorization_code"`
	Code        string    `json:"code" validate:"required"`
	RedirectUri string    `json:"redirect_uri" validate:"required,uri"`
	ClientID    uuid.UUID `json:"client_id" validate:"required"`

	AppSecret    string `json:"client_secret"`
	CodeVerifier string `json:"code_verifier"`
}

func (c GrantCredentials) IsValid() error {
	errs := make(map[string]string)

	if c.GrantType == "authorization_code" && c.AppSecret == "" && c.CodeVerifier == "" {
		errs["app_secret_or_code_verifier"] = "either app_secret or code_verifier must be provided"
	}

	if len(errs) != 0 {
		return apperr.NewMapError(errs)
	}

	return nil
}

func (c GrantCredentials) IsConfidentialClient() bool {
	return c.AppSecret != ""
}

func (c GrantCredentials) IsPKCE() bool {
	return c.CodeVerifier != ""
}
