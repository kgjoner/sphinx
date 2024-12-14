package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
	"github.com/kgjoner/cornucopia/utils/structop"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/config/errcode"
)

type Application struct {
	InternalId int       `json:"-"`
	Id         uuid.UUID `json:"-" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Grantings  []string  `json:"grantings"`

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
	Grantings           []string `json:"grantings"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
}

func NewApplication(f *ApplicationCreationFields, actor Account) (*Application, error) {
	actorApp := actor.AuthedSession.Application
	if !actorApp.isRoot() || !actor.HasRole(actorApp, RoleValues.ADMIN) {
		return nil, normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	now := time.Now()
	created := &Application{
		Id:        uuid.New(),
		Name:      f.Name,
		Grantings: f.Grantings,

		Secret:              pwdgen.Generate(42, "lower", "upper", "number"),
		AllowedRedirectUris: f.AllowedRedirectUris,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, validator.Validate(created)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (a *Application) Authenticate(secret string) error {
	if a.Secret != secret {
		return normalizederr.NewFatalUnauthorizedError("invalid credentials", errcode.InvalidAccess)
	}

	a.HasValidCredentials = true
	return nil
}

func (a Application) IsAuthenticated() bool {
	return a.HasValidCredentials
}

type ApplicationEditableFields struct {
	Name                string   `json:"name"`
	Grantings           []string `json:"grantings"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
}

func (a *Application) Edit(f *ApplicationEditableFields, actor Account) error {
	actorApp := actor.AuthedSession.Application
	if !actorApp.isRoot() || !actor.HasRole(actorApp, RoleValues.ADMIN) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	structop.New(a).Update(f)
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

func (a Application) isRoot() bool {
	return a.Id.String() == config.Env.ROOT_APP_TOKEN
}
