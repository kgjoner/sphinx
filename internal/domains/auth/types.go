package auth

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/config/errcode"
)

/* ==============================================================================
	AccountCodeKind
============================================================================== */

type AccountCodeKind string

type accountCodeKind struct {
	EMAIL_VERIFICATION AccountCodeKind
	PHONE_VERIFICATION AccountCodeKind
	PASSWORD_RESET     AccountCodeKind
}

func (s AccountCodeKind) Enumerate() any {
	return accountCodeKind{
		"email_verification",
		"phone_verification",
		"password_reset",
	}
}

var AccountCodeKindValues = AccountCodeKind.Enumerate("").(accountCodeKind)

/* ==============================================================================
	Roles
============================================================================== */

type Role string

type role struct {
	ADMIN Role
	STAFF Role
}

func (s Role) Enumerate() any {
	return role{
		"ADMIN",
		"STAFF",
	}
}

var RoleValues = Role.Enumerate("").(role)

/* ==============================================================================
	Auth Token
============================================================================== */

type authToken struct {
	jwt.Token
	Claims       jwtClaims
	signedString string
}

type authTokenCreationFields struct {
	Account   Account
	SessionId uuid.UUID
	IsRefresh bool
}

func newAuthToken(f authTokenCreationFields) (*authToken, error) {
	s := f.Account.session(f.SessionId)
	if s == nil {
		return nil, normalizederr.NewRequestError("Account and session do not match.")
	}

	now := time.Now()
	var kind string
	var duration time.Duration
	if f.IsRefresh {
		kind = "refresh"
		duration = time.Second * time.Duration(config.Env.JWT.REFRESH_LIFETIME_IN_SEC)
	} else {
		kind = "access"
		duration = time.Second * time.Duration(config.Env.JWT.ACCESS_LIFETIME_IN_SEC)
	}

	claims := jwtClaims{
		f.Account.Id,
		s.Application.Id,
		now.Unix(),
		now.Add(duration).Unix(),
		s.Id,
		kind,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenAsSignedString, err := token.SignedString([]byte(config.Env.JWT.SECRET))
	if err != nil {
		return nil, err
	}

	return &authToken{*token, claims, tokenAsSignedString}, nil
}

func ParseAuthToken(str string) (*authToken, error) {
	token, err := jwt.Parse(str, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Env.JWT.SECRET), nil
	})
	if err != nil {
		return nil, normalizederr.NewUnauthorizedError(err.Error(), errcode.InvalidAccess)
	}

	if !token.Valid {
		return nil, normalizederr.NewUnauthorizedError("Invalid authToken", errcode.InvalidAccess)
	}

	var claims jwtClaims
	ms, _ := json.Marshal(token.Claims)
	err = json.Unmarshal(ms, &claims)
	if err != nil {
		return nil, normalizederr.NewUnauthorizedError("Badly formatted authToken", errcode.InvalidAccess)
	}

	return &authToken{*token, claims, str}, nil
}

func (t authToken) IsExpired() bool {
	now := time.Now()
	return time.Unix(t.Claims.Exp, 0).Before(now)
}

func (t authToken) IsRefresh() bool {
	return t.Claims.Kind == "refresh"
}

func (t authToken) String() string {
	return t.signedString
}

type jwtClaims struct {
	Sub       uuid.UUID `json:"sub"`
	Aud       uuid.UUID `json:"aud"`
	Iat       int64     `json:"iat"`
	Exp       int64     `json:"exp"`
	SessionId uuid.UUID `json:"sessionId"`
	Kind      string    `json:"kind" validate:"oneof=refresh access"`
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
