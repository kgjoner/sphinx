package tokens

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

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

/* ==========================================================================
	TokenProvider
========================================================================== */

func (p *JWTProvider) Generate(sub auth.Subject) (*auth.Tokens, error) {
	now := time.Now()

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
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSignedToken, err := accessToken.SignedString([]byte(p.secret))
	if err != nil {
		return nil, err
	}

	refreshDuration := time.Second * time.Duration(p.refreshTokenLifetimeInSec)
	refreshClaims := jwtClaims{
		Sub:       sub.ID,
		Aud:       sub.AudienceID,
		SessionID: sub.SessionID,
		Iat:       now.Unix(),
		Exp:       now.Add(refreshDuration).Unix(),
		Intent:    auth.IntentRefresh,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSignedToken, err := refreshToken.SignedString([]byte(p.secret))
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
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(p.secret), nil
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

	var email htypes.Email
	if claims.Email != "" {
		email, err = htypes.ParseEmail(claims.Email)
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
