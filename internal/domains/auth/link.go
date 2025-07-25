package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/sliceman"
	"github.com/kgjoner/sphinx/internal/common/errcode"
)

type Link struct {
	InternalId  int         `json:"-"`
	Id          uuid.UUID   `json:"id" validate:"required"`
	AccountId   int         `json:"-" validate:"required"`
	Application Application `json:"application" validate:"required"`
	Roles       []Role      `json:"roles"`
	HasConsent  bool        `json:"hasConsent"`

	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

func newLink(acc *Account, app Application) *Link {
	now := time.Now()
	consent := &Link{
		Id:          uuid.New(),
		AccountId:   acc.InternalId,
		Application: app,
		HasConsent:    true,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return consent
}

/* ==============================================================================
	METHODS
============================================================================== */

func (c Link) hasRole(roles ...Role) bool {
	for _, existingRole := range c.Roles {
		for _, allowedRole := range roles {
			if existingRole == allowedRole {
				return true
			}
		}
	}
	return false
}

func (c *Link) addRole(r Role) error {
	if sliceman.IndexOf(c.Application.PossibleRoles, r) == -1 {
		return normalizederr.NewRequestError("Application does not support the desired role.", errcode.InvalidRole)
	}

	if sliceman.IndexOf(c.Roles, r) != -1 {
		return normalizederr.NewRequestError("Role has already been added.")
	}

	c.Roles = append(c.Roles, r)
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

func (c *Link) removeRole(r Role) error {
	if sliceman.IndexOf(c.Application.PossibleRoles, r) == -1 {
		return normalizederr.NewRequestError("Application does not support the desired role.", errcode.InvalidRole)
	}

	index := sliceman.IndexOf(c.Roles, r)
	if index == -1 {
		return normalizederr.NewRequestError("Role has not been added.")
	}

	c.Roles = sliceman.Remove(c.Roles, index)
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

func (c *Link) revokeConsent() error {
	c.HasConsent = false
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

func (c *Link) restoreConsent() error {
	c.HasConsent = true
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

