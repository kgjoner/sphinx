package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/sanitizer"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
)

/* ==============================================================================
	VerificationKind
============================================================================== */

type VerificationKind string

const (
	VerificationEmail         VerificationKind = "email"
	VerificationPhone         VerificationKind = "phone"
	VerificationPasswordReset VerificationKind = "password_reset"
)

func (s VerificationKind) Enumerate() any {
	return []VerificationKind{
		VerificationEmail,
		VerificationPhone,
		VerificationPasswordReset,
	}
}

/* ==============================================================================
	Roles
============================================================================== */

type Role string

// Roles used in root application.
const (
	RoleAdmin Role = "ADMIN"
	RoleDev   Role = "DEV"
)

/* ==============================================================================
	Entry
============================================================================== */

// Represents any entry of an user: email, phone, username or document.
type Entry string

func ParseEntry(str string) (Entry, error) {
	if str == "" {
		return "", nil
	}

	if strings.Contains(str, "@") {
		email, err := htypes.ParseEmail(str)
		return Entry(email), err
	}

	if strings.HasPrefix(str, "+") {
		phone, err := htypes.ParsePhoneNumber(str)
		return Entry(phone), err
	}

	// Try to parse a document even if it does not contain a colon.
	document, err := htypes.ParseDocument(str)
	if err == nil || (strings.Contains(str, ":") || sanitizer.IsDigitOnly(str)) {
		// If it contains a colon or digit, it should be a document.
		return Entry(document), err
	}

	e := Entry(strings.ToLower(str))
	return e, e.IsValid()
}

func (e Entry) IsValid() error {
	var err error
	kind := e.Kind()
	switch kind {
	case "email":
		err = htypes.Email(string(e)).IsValid()
	case "phone":
		err = htypes.PhoneNumber(string(e)).IsValid()
	case "document":
		err = htypes.Document(string(e)).IsValid()
	case "username":
		err = validator.Validate(string(e), "wordID", "atLeastOne=letter")
	}

	if err != nil {
		return fmt.Errorf("invalid entry, identified like %v: %v", kind, err)
	}

	return nil
}

func (e Entry) Kind() string {
	str := string(e)
	switch {
	case strings.Contains(str, "@"):
		return "email"
	case strings.Contains(str, "+"):
		return "phone"
	case strings.Contains(str, ":") || sanitizer.IsDigitOnly(str):
		return "document"
	default:
		return "username"
	}
}

func (e Entry) String() string {
	return string(e)
}

func (e *Entry) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	*e, err = ParseEntry(s)
	return err
}

/* ==============================================================================
	Auth Token
============================================================================== */

type authToken struct {
	jwt.Token
	Claims       jwtClaims
	signedString string
}

type authTokenCreationFields struct {
	User      User
	SessionID uuid.UUID
	IsRefresh bool
}

func newAuthToken(f authTokenCreationFields) (*authToken, error) {
	s := f.User.session(f.SessionID)
	if s == nil {
		return nil, normalizederr.NewRequestError("User and session do not match.")
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
		f.User.ID,
		s.Application.ID,
		now.Unix(),
		now.Add(duration).Unix(),
		s.ID,
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
		msg := err.Error()
		if strings.Contains(msg, "token is expired") {
			iat, _ := token.Claims.GetIssuedAt()
			exp, _ := token.Claims.GetExpirationTime()
			diff := exp.Sub(iat.Time)
			code := errcode.ExpiredAccess
			if diff.Seconds() >= float64(config.Env.JWT.REFRESH_LIFETIME_IN_SEC) {
				code = errcode.ExpiredSession
			}
			return nil, normalizederr.NewUnauthorizedError(msg, code)
		} else {
			code := errcode.InvalidAccess
			return nil, normalizederr.NewUnauthorizedError(msg, code)
		}
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
	SessionID uuid.UUID `json:"sessionID"`
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
