package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/sliceman"
)

type Link struct {
	InternalId  int         `json:"-"`
	Id          uuid.UUID   `json:"id" validate:"required"`
	AccountId   int         `json:"account_id" validate:"required"`
	Application Application `json:"application" validate:"required"`
	Roles       []Role      `json:"roles"`
	Grantings   []string    `json:"grantings"`

	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
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
		return normalizederr.NewRequestError("Role has already been added.", "")
	}

	l.Roles = append(l.Roles, r)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) removeRole(r Role) error {
	index := sliceman.IndexOf(l.Roles, r)
	if index == -1 {
		return normalizederr.NewRequestError("Role has not been added.", "")
	}

	l.Roles = sliceman.Remove(l.Roles, index)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) addGranting(g string) error {
	if sliceman.IndexOf(l.Application.Grantings, g) == -1 {
		return normalizederr.NewRequestError("Application does not support the desired granting.", "")
	}

	if sliceman.IndexOf(l.Grantings, g) != -1 {
		return normalizederr.NewRequestError("Granting has already been added.", "")
	}

	l.Grantings = append(l.Grantings, g)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) removeGranting(g string) error {
	if sliceman.IndexOf(l.Application.Grantings, g) == -1 {
		return normalizederr.NewRequestError("Application does not support the desired granting.", "")
	}

	index := sliceman.IndexOf(l.Grantings, g)
	if index == -1 {
		return normalizederr.NewRequestError("Granting has not been added.", "")
	}

	l.Grantings = sliceman.Remove(l.Grantings, index)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}
