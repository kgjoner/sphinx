package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/cornucopia/v2/utils/pwdgen"
	"github.com/kgjoner/cornucopia/v2/utils/structop"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type Application struct {
	InternalID    int       `json:"-"`
	ID            uuid.UUID `json:"id" validate:"required"`
	Name          string    `json:"name" validate:"required"`
	PossibleRoles []Role    `json:"possibleRoles"`

	Secret              string   `json:"-" validate:"required"`
	AllowedRedirectUris []string `json:"allowedRedirectUris" validate:"uri"`

	HasValidCredentials bool `json:"-"`

	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type ApplicationCreationFields struct {
	Name                string   `json:"name" validate:"required"`
	PossibleRoles       []Role   `json:"possibleRoles"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
}

func NewApplication(f *ApplicationCreationFields, actor User) (app *Application, secret string, err error) {
	actorApp := actor.AuthedSession.Application
	if !actorApp.isRoot() || !(actor.HasRole(actorApp, RoleAdmin) || actor.HasRole(actorApp, RoleDev)) {
		return nil, "", apperr.NewForbiddenError("Does not have permission to execute this action.")
	}

	secret = generateAppSecret()
	now := time.Now()
	created := &Application{
		ID:            uuid.New(),
		Name:          f.Name,
		PossibleRoles: f.PossibleRoles,

		Secret:              hashPassword(secret),
		AllowedRedirectUris: f.AllowedRedirectUris,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, secret, validator.Validate(created)
}

func generateAppSecret() string {
	return pwdgen.GeneratePassword(42, "lower", "upper", "number")
}

/* ==============================================================================
	METHODS
============================================================================== */

type ApplicationEditableFields struct {
	Name                string   `json:"name"`
	PossibleRoles       []Role   `json:"possibleRoles"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
}

func (a *Application) Edit(f *ApplicationEditableFields, actor User) error {
	actorApp := actor.AuthedSession.Application
	if actorApp.ID != a.ID || !actor.HasRoleOnAuth(RoleAdmin) {
		return apperr.NewForbiddenError("Does not have permission to execute this action.")
	}

	structop.New(a).Update(f)
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

func (a *Application) GenerateNewSecret(actor User) (secret string, err error) {
	actorApp := actor.AuthedSession.Application
	if actorApp.ID != a.ID || !actor.HasRoleOnAuth(RoleAdmin) {
		return "", apperr.NewForbiddenError("Does not have permission to execute this action.")
	}

	secret = generateAppSecret()
	a.Secret = hashPassword(secret)
	a.UpdatedAt = time.Now()
	return secret, validator.Validate(a)
}

func (a Application) isRoot() bool {
	return a.ID.String() == config.Env.ROOT_APP_ID
}

func (a *Application) DoesSecretMatch(secret string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Secret), []byte(secret))
	return err == nil
}

func (a *Application) Authenticate(secret string) error {
	if !a.DoesSecretMatch(secret) {
		return apperr.Fatal(apperr.NewUnauthorizedError("invalid credentials", errcode.InvalidAccess))
	}

	a.HasValidCredentials = true
	return nil
}

func (a Application) IsAuthenticated() bool {
	return a.HasValidCredentials
}
