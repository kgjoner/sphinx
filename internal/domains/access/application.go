package access

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/validator"
	"github.com/kgjoner/cornucopia/v3/structop"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Application struct {
	ID            uuid.UUID `validate:"required"`
	Name          string    `validate:"required"`
	PossibleRoles []Role

	Secret              shared.HashedPassword `validate:"required"`
	AllowedRedirectUris []string              `validate:"uri"`

	HasValidCredentials bool

	CreatedAt time.Time `validate:"required"`
	UpdatedAt time.Time `validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type ApplicationCreationFields struct {
	Name                string                `json:"name" validate:"required"`
	Secret              shared.HashedPassword `json:"secret" validate:"required"`
	PossibleRoles       []Role                `json:"possibleRoles"`
	AllowedRedirectUris []string              `json:"allowedRedirectUris"`
}

func NewApplication(f *ApplicationCreationFields) (app *Application, err error) {
	now := time.Now()
	created := &Application{
		ID:            uuid.New(),
		Name:          f.Name,
		PossibleRoles: f.PossibleRoles,

		Secret:              f.Secret,
		AllowedRedirectUris: f.AllowedRedirectUris,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, validator.Validate(created)
}

/* ==============================================================================
	METHODS
============================================================================== */

type ApplicationEditableFields struct {
	Name                string   `json:"name"`
	PossibleRoles       []Role   `json:"extraRoles"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
}

func (a *Application) Edit(f *ApplicationEditableFields) error {
	structop.New(a).Update(f)
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

func (a *Application) UpdateSecret(proof shared.PasswordProof, newSecret shared.HashedPassword) error {
	if !proof.ValidFor(a.Secret) {
		return shared.ErrInvalidCredentials
	}

	a.Secret = newSecret
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

func (a Application) isRoot() bool {
	return a.ID.String() == config.Env.ROOT_APP_ID
}

/* ==============================================================================
	VIEWS
============================================================================== */

type ApplicationView struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	PossibleRoles       []Role    `json:"extraRoles"`
	AllowedRedirectUris []string  `json:"allowedRedirectUris"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (a Application) View() ApplicationView {
	return ApplicationView{
		ID:                  a.ID,
		Name:                a.Name,
		PossibleRoles:       a.PossibleRoles,
		AllowedRedirectUris: a.AllowedRedirectUris,

		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

type ApplicationLeanView struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (a Application) LeanView() ApplicationLeanView {
	return ApplicationLeanView{
		ID:   a.ID,
		Name: a.Name,
	}
}
