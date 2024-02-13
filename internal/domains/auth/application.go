package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/structop"
	"github.com/kgjoner/sphinx/internal/config"
)

type Application struct {
	InternalId int       `json:"-"`
	Id         uuid.UUID `json:"-" validator:"required"`
	Name       string    `json:"name" validator:"required"`
	Grantings  []string  `json:"grantings"`

	CreatedAt time.Time `json:"createdAt" validator:"required"`
	UpdatedAt time.Time `json:"updatedAt" validator:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type ApplicationCreationFields struct {
	Name      string
	Grantings []string
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

		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, validator.Validate(created)
}

/* ==============================================================================
	METHODS
============================================================================== */

type ApplicationEditableFields struct {
	Name      string
	Grantings []string
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
	return a.Id.String() == config.Environment.ROOT_APP_TOKEN
}
