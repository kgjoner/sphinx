package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
)

// Represents an OAuth 2.0 authorization grant. It should have an ephemeral duration,
// typically persisted in cache with a TTL. All fields are kept in JSON serialization
// to facilitate caching as JSON; be careful when sending them out.
type AuthorizationGrant struct {
	Type        string    `json:"grantType" validate:"required,oneof=authorization_code"`
	LinkId      uuid.UUID `json:"linkId" validate:"required"`
	ExpiresAt   time.Time `json:"expiresAt" validate:"required"`
	RedirectUri string    `json:"redirectUri" validate:"required,url"`
	IsUsed      bool      `json:"isUsed"`

	Code          string `json:"code" validate:"required"`
	CodeChallenge string `json:"codeChallenge"`
	// Plain is not recommended, but part of spec. Default is S256.
	CodeChallengeMethod string `json:"codeChallengeMethod" validate:"oneof=S256 plain"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type AuthorizationGrantCreationFields struct {
	Type                string    `json:"grant_type" validate:"required,oneof=authorization_code"`
	ClientId            uuid.UUID `json:"client_id" validate:"required"` // Application ID
	RedirectUri         string    `json:"redirect_uri" validate:"required,url"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method" validate:"oneof=S256 plain"`
}

func newAuthorizationGrant(acc Account, g *AuthorizationGrantCreationFields) (*AuthorizationGrant, error) {
	link := acc.link(g.ClientId)
	if link == nil || !link.HasConsent {
		return nil, normalizederr.NewForbiddenError("user has not consented to this application", errcode.NoConsent)
	}

	// Validate redirect URI is allowed by application
	var hasFound bool
	for _, uri := range link.Application.AllowedRedirectUris {
		if uri == g.RedirectUri {
			hasFound = true
			break
		}
	}
	if !hasFound {
		return nil, normalizederr.NewRequestError("Invalid redirect_uri")
	}

	var code string
	switch g.Type {
	case "authorization_code":
		code = pwdgen.Generate(42, "lower", "upper", "number")
		if g.CodeChallenge != "" && g.CodeChallengeMethod == "" {
			g.CodeChallengeMethod = "S256" // Default to S256 if not specified
		}
	}

	now := time.Now()
	grant := &AuthorizationGrant{
		Code:                code,
		LinkId:              link.Id,
		ExpiresAt:           now.Add(time.Second * time.Duration(config.Env.AUTH_GRANT_LIFETIME_IN_SEC)),
		RedirectUri:         g.RedirectUri,
		CodeChallenge:       g.CodeChallenge,
		CodeChallengeMethod: g.CodeChallengeMethod,
	}

	return grant, validator.Validate(grant)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (g AuthorizationGrant) IsValid() error {
	errs := make(map[string]error)

	if g.Type == "authorization_code" && g.CodeChallengeMethod != "" && g.CodeChallenge == "" {
		errs["code_challenge"] = fmt.Errorf("code_challenge is required when code_challenge_method is set")
	}

	if len(errs) != 0 {
		return normalizederr.NewValidationErrorFromMap(errs)
	}

	return nil
}

func (g *AuthorizationGrant) isExpired() bool {
	return g.ExpiresAt.Before(time.Now())
}

func (g *AuthorizationGrant) isActive() bool {
	return !g.IsUsed && !g.isExpired()
}

// HANDLE GRANT VALIDATION AND CONSUMING

type GrantCredentials struct {
	GrantType   string    `json:"grant_type" validate:"required,oneof=authorization_code"`
	Code        string    `json:"code" validate:"required"`
	RedirectUri string    `json:"redirect_uri" validate:"required,url"`
	ClientId    uuid.UUID `json:"client_id" validate:"required"` // Application ID

	AppSecret    string `json:"client_secret"` // For confidential clients
	CodeVerifier string `json:"code_verifier"` // For PKCE (public clients)
}

func (c GrantCredentials) IsValid() error {
	errs := make(map[string]error)

	if c.GrantType == "authorization_code" && c.AppSecret == "" && c.CodeVerifier == "" {
		errs["app_secret_or_code_verifier"] = fmt.Errorf("either app_secret or code_verifier must be provided")
	}

	if len(errs) != 0 {
		return normalizederr.NewValidationErrorFromMap(errs)
	}

	return nil
}

// Checks if grant credentials are valid, returning an error otherwise. Either way consumes the grant marking it as used.
func (g *AuthorizationGrant) use(acc Account, credentials *GrantCredentials) error {
	defer func() { g.IsUsed = true }()

	err := validator.Validate(credentials)
	if err != nil {
		return err
	}

	// Check link mataches
	link := acc.link(credentials.ClientId)
	if link == nil || link.Id != g.LinkId {
		return normalizederr.NewUnauthorizedError("Invalid consent", errcode.InvalidCredentials)
	}

	// Check basic grant validity
	if !g.isActive() {
		if g.IsUsed {
			return normalizederr.NewUnauthorizedError("Authorization code has already been used.", errcode.InvalidCredentials)
		}
		return normalizederr.NewUnauthorizedError("OAuth code has expired.", errcode.InvalidCredentials)
	}

	// Check simple credentials matches
	if credentials.GrantType != g.Type || credentials.Code != g.Code || credentials.RedirectUri != g.RedirectUri || credentials.ClientId != link.Application.Id {
		return normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	switch g.Type {
	case "authorization_code":
		err = g.validateAuthorizationCode(link, credentials)
		if err != nil {
			return err
		}
	}

	// Check consent is still valid
	if !link.HasConsent {
		return normalizederr.NewForbiddenError("user has revoked consent to application", errcode.RevokedConsent)
	}

	return validator.Validate(g)
}

func (g AuthorizationGrant) validateAuthorizationCode(link *Link, credentials *GrantCredentials) error {
	// Determine client type to validate accordingly
	isConfidentialClient := g.CodeChallenge == ""

	// 1. Confidential client using client_secret
	if isConfidentialClient {
		if !link.Application.DoesSecretMatch(credentials.AppSecret) {
			return normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
		}

		return nil
	}

	// 2. Public client using PKCE
	if credentials.CodeVerifier == "" {
		return normalizederr.NewUnauthorizedError("code_verifier required for PKCE flow", errcode.InvalidCredentials)
	}

	var isCodeVerifierValid bool
	switch g.CodeChallengeMethod {
	case "S256":
		hash := sha256.Sum256([]byte(credentials.CodeVerifier))
		// Base64 URL encode (without padding)
		computed := strings.TrimRight(base64.URLEncoding.EncodeToString(hash[:]), "=")
		isCodeVerifierValid = computed == g.CodeChallenge
	case "plain":
		isCodeVerifierValid = credentials.CodeVerifier == g.CodeChallenge
	}

	if !isCodeVerifierValid {
		return normalizederr.NewUnauthorizedError("Invalid code_verifier", errcode.InvalidCredentials)
	}

	return nil
}
