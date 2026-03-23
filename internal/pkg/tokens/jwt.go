package tokens

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type JWTProvider struct {
	accessTokenLifetimeInSec  int
	refreshTokenLifetimeInSec int
	secret                    string     // Legacy HS256 secret (deprecated)
	keysGetter                KeysGetter // For RS256 key management
}

// NewJWTProvider creates a JWT provider with HS256 support (legacy).
// Deprecated: Use NewJWTProviderWithKeyManager for RS256 support.
func NewJWTProvider(secret string, accessLifetime int, refreshLifetime int) *JWTProvider {
	return &JWTProvider{
		accessTokenLifetimeInSec:  accessLifetime,
		refreshTokenLifetimeInSec: refreshLifetime,
		secret:                    secret,
		keysGetter:                nil,
	}
}

// NewJWTProviderWithKeyPair creates a JWT provider with RS256 key rotation support.
func NewJWTProviderWithKeyPair(
	keysGetter KeysGetter,
	secret string, // For legacy HS256 support
	accessLifetime int,
	refreshLifetime int,

) *JWTProvider {
	return &JWTProvider{
		accessTokenLifetimeInSec:  accessLifetime,
		refreshTokenLifetimeInSec: refreshLifetime,
		secret:                    secret, // For legacy validation
		keysGetter:                keysGetter,
	}
}

/* ==========================================================================
	TokenProvider
========================================================================== */

func (p *JWTProvider) Generate(sub auth.Subject) (*auth.Tokens, error) {
	now := time.Now()

	// Determine signing method and key
	var signingMethod jwt.SigningMethod
	var signingKey interface{}
	var kid string
	var err error

	if p.keysGetter != nil {
		// Use RS256 with key provisioner
		signingKey, kid, err = p.keysGetter.CurrentSigningKey()
		if err != nil {
			return nil, err
		}
		signingMethod = jwt.SigningMethodRS256
	} else {
		// Fallback to legacy HS256
		signingMethod = jwt.SigningMethodHS256
		signingKey = []byte(p.secret)
		kid = "hs256-legacy"
	}

	// Generate access token
	accessDuration := time.Second * time.Duration(p.accessTokenLifetimeInSec)
	accessClaims := jwtClaims{
		Sub:       sub.ID,
		Aud:       sub.AudienceID,
		Iat:       now.Unix(),
		Exp:       now.Add(accessDuration).Unix(),
		Intent:    auth.IntentAccess,
		Email:     sub.Email.String(),
		Name:      sub.Name,
		Roles:     sub.Roles,
		SessionID: sub.SessionID,
	}
	accessToken := jwt.NewWithClaims(signingMethod, accessClaims)
	accessToken.Header["kid"] = kid // Add key ID to header
	accessSignedToken, err := accessToken.SignedString(signingKey)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshDuration := time.Second * time.Duration(p.refreshTokenLifetimeInSec)
	refreshClaims := jwtClaims{
		Sub:       sub.ID,
		Aud:       sub.AudienceID,
		SessionID: sub.SessionID,
		Iat:       now.Unix(),
		Exp:       now.Add(refreshDuration).Unix(),
		Intent:    auth.IntentRefresh,
	}
	refreshToken := jwt.NewWithClaims(signingMethod, refreshClaims)
	refreshToken.Header["kid"] = kid // Add key ID to header
	refreshSignedToken, err := refreshToken.SignedString(signingKey)
	if err != nil {
		return nil, err
	}

	return &auth.Tokens{
		AccessToken:  accessSignedToken,
		RefreshToken: refreshSignedToken,
		ExpiresIn:    int(accessDuration.Seconds()),
	}, nil
}

func (p *JWTProvider) Validate(signedToken string) (*auth.Subject, auth.Intent, error) {
	// Parse token to extract kid from header
	var kid string
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		// Extract kid from header
		if kidHeader, ok := t.Header["kid"].(string); ok {
			kid = kidHeader
		}

		// Validate algorithm
		if t.Method.Alg() == "HS256" {
			// Legacy HS256 validation
			if kid == "" || kid == "hs256-legacy" {
				return []byte(p.secret), nil
			}
			return nil, errors.New("invalid HS256 key ID")
		}

		if t.Method.Alg() == "RS256" {
			// RS256 validation with key manager
			if p.keysGetter == nil {
				return nil, errors.New("RS256 not supported without key manager")
			}
			if kid == "" {
				return nil, errors.New("missing kid in RS256 token")
			}

			publicKey, algorithm, err := p.keysGetter.PublicKeyByKID(kid)
			if err != nil {
				return nil, err
			}
			if algorithm != "RS256" {
				return nil, errors.New("key algorithm mismatch")
			}
			return publicKey, nil
		}

		return nil, errors.New("unsupported signing method")
	})

	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "token is expired") {
			iat, _ := token.Claims.GetIssuedAt()
			exp, _ := token.Claims.GetExpirationTime()
			diff := exp.Sub(iat.Time)
			err = auth.ErrExpiredAccess
			if diff.Seconds() >= float64(p.refreshTokenLifetimeInSec) {
				err = auth.ErrExpiredSession
			}
			return nil, "", err
		} else {
			return nil, "", auth.ErrInvalidAccess
		}
	}

	if !token.Valid {
		return nil, "", auth.ErrInvalidAccess
	}

	var claims jwtClaims
	ms, _ := json.Marshal(token.Claims)
	err = json.Unmarshal(ms, &claims)
	if err != nil {
		return nil, "", auth.ErrInvalidAccess
	}

	var email prim.Email
	if claims.Email != "" {
		email, err = prim.ParseEmail(claims.Email)
		if err != nil {
			return nil, "", auth.ErrInvalidAccess
		}
	}

	return &auth.Subject{
		ID:         claims.Sub,
		Kind:       shared.KindUser,
		Email:      email,
		Name:       claims.Name,
		AudienceID: claims.Aud,
		Roles:      claims.Roles,
		SessionID:  claims.SessionID,
	}, claims.Intent, nil
}

/* ==========================================================================
	Token Claims
========================================================================== */

type jwtClaims struct {
	Sub    uuid.UUID   `json:"sub"`
	Aud    uuid.UUID   `json:"aud"`
	Iat    int64       `json:"iat"`
	Exp    int64       `json:"exp"`
	Intent auth.Intent `json:"itn"`

	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles"`
	SessionID uuid.UUID `json:"sid"`
}

func (c jwtClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Aud.String()}, nil
}
func (c jwtClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: time.Unix(c.Exp, 0)}, nil
}
func (c jwtClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: time.Unix(c.Iat, 0)}, nil
}
func (c jwtClaims) GetIssuer() (string, error) {
	return "", nil
}
func (c jwtClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{}, nil
}
func (c jwtClaims) GetSubject() (string, error) {
	return c.Sub.String(), nil
}
