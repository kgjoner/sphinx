package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/structop"
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

func NewApplication(f *ApplicationCreationFields) (*Application, error) {
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

func (e *Application) Edit(f *ApplicationEditableFields) error {
	structop.New(e).Update(f)
	e.UpdatedAt = time.Now()
	return validator.Validate(e)
}
