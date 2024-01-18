package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config"
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

func newAuthToken(a Account, sessionId uuid.UUID, kindFlag ...string) (*authToken, error) {
	s := a.Session(sessionId)
	if s == nil {
		return nil, normalizederr.NewRequestError("Account and session do not match.", "")
	}

	now := time.Now()
	var kind string
	var duration time.Duration
	if len(kindFlag) >= 1 && kindFlag[0] == "refresh" {
		kind = "refresh"
		duration = time.Second * time.Duration(config.Environment.JWT.REFRESH_LIFE_TIME_IN_SEC)
	} else {
		kind = "access"
		duration = time.Second * time.Duration(config.Environment.JWT.ACCESS_LIFE_TIME_IN_SEC)
	}

	claims := jwtClaims{
		a.Id,
		s.Application.Id,
		now,
		now.Add(duration),
		s.Id,
		kind,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenAsSignedString, err := token.SignedString([]byte(config.Environment.JWT.SECRET))
	if err != nil {
		return nil, err
	}

	return &authToken{*token, claims, tokenAsSignedString}, nil
}

func ParseAuthToken(str string) (*authToken, error) {
	token, err := jwt.Parse(str, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Environment.JWT.SECRET), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, normalizederr.NewRequestError("Invalid authToken", "")
	}

	claims, ok := token.Claims.(jwtClaims)
	if !ok {
		return nil, normalizederr.NewRequestError("Badly formatted authToken", "")
	}

	return &authToken{*token, claims, str}, nil
}

func (t authToken) IsExpired() bool {
	now := time.Now()
	return t.Claims.Exp.Before(now)
}

func (t authToken) String() string {
	return t.signedString
}

type jwtClaims struct {
	Sub       uuid.UUID `json:"sub"`
	Aud       uuid.UUID `json:"aud"`
	Iat       time.Time `json:"iat"`
	Exp       time.Time `json:"exp"`
	SessionId uuid.UUID `json:"sessionId"`
	Kind      string    `json:"kind" validate:"oneof=refresh access"`
}

func (c jwtClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Aud.String()}, nil
}
func (c jwtClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: c.Exp}, nil
}
func (c jwtClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: c.Iat}, nil
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
