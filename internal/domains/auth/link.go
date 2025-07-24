package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
	"github.com/kgjoner/cornucopia/utils/sliceman"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
)

type Link struct {
	InternalId  int         `json:"-"`
	Id          uuid.UUID   `json:"id" validate:"required"`
	AccountId   int         `json:"-" validate:"required"`
	Application Application `json:"application" validate:"required"`
	Roles       []Role      `json:"roles"`
	Grantings   []string    `json:"grantings"`

	OAuthCode                string          `json:"-"`
	OAuthExpiresAt           htypes.NullTime `json:"-"`
	OAuthCodeChallenge       string          `json:"-"`
	OAuthCodeChallengeMethod string          `json:"-" validate:"oneof=S256 plain"`

	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

func newLink(acc *Account, app Application) *Link {
	now := time.Now()
	link := &Link{
		Id:          uuid.New(),
		AccountId:   acc.InternalId,
		Application: app,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return link
}

/* ==============================================================================
	METHODS
============================================================================== */

// Save code and set an expiration time for it
func (l *Link) initOAuth(codeChallenge, codeChallengeMethod string, code ...string) error {
	var realCode string
	if len(code) > 0 {
		realCode = code[0]
	} else {
		realCode = pwdgen.Generate(42, "lower", "upper", "number")
	}

	if codeChallenge != "" {
		if codeChallengeMethod == "" {
			codeChallengeMethod = "S256" // Default to S256 if not specified
		}
	}

	l.OAuthCode = realCode
	l.OAuthExpiresAt = htypes.NullTime{Time: time.Now().Add(time.Second * time.Duration(config.Env.OAUTH_LIFETIME_IN_SEC))}
	l.OAuthCodeChallenge = codeChallenge
	l.OAuthCodeChallengeMethod = codeChallengeMethod
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

// validatePKCE validates the code_verifier against the stored code_challenge
func (l *Link) validatePKCE(codeVerifier string) bool {
	if l.OAuthCodeChallenge == "" {
		return false // No PKCE challenge stored
	}

	if l.OAuthCodeChallengeMethod == "S256" {
		// Hash the code_verifier with SHA256
		hash := sha256.Sum256([]byte(codeVerifier))
		// Base64 URL encode (without padding)
		computed := strings.TrimRight(base64.URLEncoding.EncodeToString(hash[:]), "=")
		return computed == l.OAuthCodeChallenge
	} else if l.OAuthCodeChallengeMethod == "plain" {
		// Plain text comparison (not recommended, but part of spec)
		return codeVerifier == l.OAuthCodeChallenge
	}

	return false
}

type OAuthAuthenticateFields struct {
	GrantType   string    `json:"grant_type" validate:"required,oneof=authorization_code"`
	Code        string    `json:"code" validate:"required"`
	RedirectUri string    `json:"redirect_uri" validate:"required,url"`
	AppId       uuid.UUID `json:"client_id" validate:"required"`

	AppSecret    string `json:"client_secret"` // For confidential clients
	CodeVerifier string `json:"code_verifier"` // For PKCE (public clients)
}

func (f OAuthAuthenticateFields) IsValid() error {
	errs := make(map[string]error)

	if f.GrantType == "authorization_code" && f.AppSecret == "" && f.CodeVerifier == "" {
		errs["app_secret_or_code_verifier"] = fmt.Errorf("either app_secret or code_verifier must be provided")
	}

	if len(errs) != 0 {
		return normalizederr.NewValidationErrorFromMap(errs)
	}

	return nil
}

// Return nil if the pair code/secret matches or error otherwise. In either case, it clears oauth data.
// It supports both PKCE (for public clients) and client_secret (for confidential clients)
func (l *Link) useOAuth(f *OAuthAuthenticateFields) error {
	var err error = nil

	var hasFound bool
	for _, uri := range l.Application.AllowedRedirectUris {
		if uri == f.RedirectUri {
			hasFound = true
			break
		}
	}
	if !hasFound {
		l.clearOAuthData()
		return normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	// Check if code matches
	if f.Code != l.OAuthCode {
		err = normalizederr.NewUnauthorizedError("Invalid credentials.", errcode.InvalidCredentials)
	} else if l.OAuthExpiresAt.Before(time.Now()) {
		err = normalizederr.NewRequestError("OAuth code has expired.")
	} else {
		// Determine client type and validate accordingly
		isPKCEClient := l.OAuthCodeChallenge != ""
		isConfidentialClient := f.AppSecret != ""

		if isPKCEClient {
			// Public client using PKCE
			if f.CodeVerifier == "" {
				err = normalizederr.NewUnauthorizedError("code_verifier required for PKCE flow.", errcode.InvalidCredentials)
			} else if !l.validatePKCE(f.CodeVerifier) {
				err = normalizederr.NewUnauthorizedError("Invalid code_verifier.", errcode.InvalidCredentials)
			}
		} else if isConfidentialClient {
			// Confidential client using client_secret
			if !l.Application.DoesSecretMatch(f.AppSecret) {
				err = normalizederr.NewUnauthorizedError("Invalid credentials.", errcode.InvalidCredentials)
			}
		} else {
			// Neither PKCE nor client_secret provided
			err = normalizederr.NewUnauthorizedError("Either client_secret or code_verifier must be provided.", errcode.InvalidCredentials)
		}
	}

	l.clearOAuthData()
	if err != nil {
		return err
	}
	return validator.Validate(l)
}

func (l *Link) clearOAuthData() {
	l.OAuthCode = ""
	l.OAuthExpiresAt = htypes.NullTime{}
	l.OAuthCodeChallenge = ""
	l.OAuthCodeChallengeMethod = ""
	l.UpdatedAt = time.Now()
}

func (l Link) hasRole(roles ...Role) bool {
	for _, existingRole := range l.Roles {
		for _, allowedRole := range roles {
			if existingRole == allowedRole {
				return true
			}
		}
	}
	return false
}

func (l *Link) addRole(r Role) error {
	if sliceman.IndexOf(l.Roles, r) != -1 {
		return normalizederr.NewRequestError("Role has already been added.")
	}

	l.Roles = append(l.Roles, r)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) removeRole(r Role) error {
	index := sliceman.IndexOf(l.Roles, r)
	if index == -1 {
		return normalizederr.NewRequestError("Role has not been added.")
	}

	l.Roles = sliceman.Remove(l.Roles, index)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l Link) hasGranting(grantings ...string) bool {
	for _, existingGranting := range l.Grantings {
		for _, allowedGranting := range grantings {
			if existingGranting == allowedGranting {
				return true
			}
		}
	}
	return false
}

func (l *Link) addGranting(g string) error {
	if sliceman.IndexOf(l.Application.Grantings, g) == -1 {
		return normalizederr.NewRequestError("Application does not support the desired granting.", errcode.InvalidGranting)
	}

	if sliceman.IndexOf(l.Grantings, g) != -1 {
		return normalizederr.NewRequestError("Granting has already been added.")
	}

	l.Grantings = append(l.Grantings, g)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) removeGranting(g string) error {
	if sliceman.IndexOf(l.Application.Grantings, g) == -1 {
		return normalizederr.NewRequestError("Application does not support the desired granting.", errcode.InvalidGranting)
	}

	index := sliceman.IndexOf(l.Grantings, g)
	if index == -1 {
		return normalizederr.NewRequestError("Granting has not been added.")
	}

	l.Grantings = sliceman.Remove(l.Grantings, index)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}
