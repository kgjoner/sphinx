package auth

import (
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

	OAuthCode      string          `json:"-"`
	OAuthExpiresAt htypes.NullTime `json:"-"`

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
func (l Link) initOAuth(code ...string) error {
	var realCode string
	if len(code) > 0 {
		realCode = code[0]
	} else {
		realCode = pwdgen.Generate(42, "lower", "upper", "number")
	}

	l.OAuthCode = realCode
	l.OAuthExpiresAt = htypes.NullTime{Time: time.Now().Add(time.Second * time.Duration(config.Env.OAUTH_LIFETIME_IN_SEC))}
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

// Return nil if the pair code/secret matches or error otherwise. In either case, it clears oauth data.
func (l Link) useOAuth(code string, appSecret string) error {
	var err error = nil
	if !l.Application.DoesSecretMatch(appSecret) {
		err = normalizederr.NewUnauthorizedError("Invalid credentials.", errcode.InvalidCredentials)
	} else if code != l.OAuthCode {
		err = normalizederr.NewUnauthorizedError("Invalid credentials.", errcode.InvalidCredentials)
	} else if l.OAuthExpiresAt.Before(time.Now()) {
		err = normalizederr.NewRequestError("OAuth code has expired.")
	}

	l.OAuthCode = ""
	l.OAuthExpiresAt = htypes.NullTime{}
	l.UpdatedAt = time.Now()

	if err != nil {
		return err
	}
	return validator.Validate(l)
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
