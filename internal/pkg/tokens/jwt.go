package tokens

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/pkg/errcode"
)

/* ==========================================================================
	PROVIDER IMPLEMENTATION
========================================================================== */

type JWTProvider struct {
	accessTokenLifetimeInSec  int
	refreshTokenLifetimeInSec int
	secret                    string
}

func NewJWTProvider(secret string, accessLifetime int, refreshLifetime int) *JWTProvider {
	return &JWTProvider{
		accessTokenLifetimeInSec:  accessLifetime,
		refreshTokenLifetimeInSec: refreshLifetime,
		secret:                    secret,
	}
}

func (p *JWTProvider) Generate(sub Subject) (*Tokens, error) {
	now := time.Now()

	accessDuration := time.Second * time.Duration(p.accessTokenLifetimeInSec)
	accessClaims := jwtClaims{
		sub.UserID,
		sub.ApplicationID,
		now.Unix(),
		now.Add(accessDuration).Unix(),
		IntentAccess,
		sub.UserEmail.String(),
		sub.UserName,
		sub.Roles,
		sub.SessionID,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSignedToken, err := accessToken.SignedString([]byte(p.secret))
	if err != nil {
		return nil, err
	}

	refreshDuration := time.Second * time.Duration(p.refreshTokenLifetimeInSec)
	refreshClaims := jwtClaims{
		Sub:       sub.UserID,
		SessionID: sub.SessionID,
		Iat:       now.Unix(),
		Exp:       now.Add(refreshDuration).Unix(),
		Intent:    IntentRefresh,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSignedToken, err := refreshToken.SignedString([]byte(p.secret))
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessSignedToken,
		RefreshToken: refreshSignedToken,
		ExpiresIn:    int(accessDuration.Seconds()),
	}, nil
}

func (p *JWTProvider) Validate(signedToken string) (*Subject, Intent, error) {
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(p.secret), nil
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "token is expired") {
			iat, _ := token.Claims.GetIssuedAt()
			exp, _ := token.Claims.GetExpirationTime()
			diff := exp.Sub(iat.Time)
			code := errcode.ExpiredAccess
			if diff.Seconds() >= float64(p.refreshTokenLifetimeInSec) {
				code = errcode.ExpiredSession
			}
			return nil, "", apperr.NewUnauthorizedError(msg, code)
		} else {
			code := errcode.InvalidAccess
			return nil, "", apperr.NewUnauthorizedError(msg, code)
		}
	}

	if !token.Valid {
		return nil, "", apperr.NewUnauthorizedError("Invalid jwtToken", errcode.InvalidAccess)
	}

	var claims jwtClaims
	ms, _ := json.Marshal(token.Claims)
	err = json.Unmarshal(ms, &claims)
	if err != nil {
		return nil, "", apperr.NewUnauthorizedError("Badly formatted jwtToken", errcode.InvalidAccess)
	}

	var email htypes.Email
	if claims.Email != "" {
		email, err = htypes.ParseEmail(claims.Email)
		if err != nil {
			return nil, "", apperr.NewUnauthorizedError("Badly formatted email in jwtToken", errcode.InvalidAccess)
		}
	}

	return &Subject{
		SessionID:     claims.SessionID,
		UserID:        claims.Sub,
		UserEmail:     email,
		UserName:      claims.Name,
		ApplicationID: claims.Aud,
		Roles:         claims.Roles,
	}, claims.Intent, nil
}

/* ==========================================================================
	TOKEN CLAIMS
========================================================================== */

type jwtClaims struct {
	Sub    uuid.UUID `json:"sub"`
	Aud    uuid.UUID `json:"aud"`
	Iat    int64     `json:"iat"`
	Exp    int64     `json:"exp"`
	Intent Intent    `json:"itn"`

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
