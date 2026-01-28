package access

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/cornucopia/v2/utils/sliceman"
)

type Link struct {
	InternalID  int
	ID          uuid.UUID   `validate:"required"`
	UserID      uuid.UUID   `validate:"required"`
	Application Application `validate:"required"`
	Roles       []Role
	HasConsent  bool

	CreatedAt time.Time `validate:"required"`
	UpdatedAt time.Time `validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

func (a *Application) NewLink(userID uuid.UUID) (*Link, error) {
	now := time.Now()
	consent := &Link{
		ID:          uuid.New(),
		UserID:      userID,
		Application: *a,
		HasConsent:  true,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return consent, nil
}

/* ==============================================================================
	METHODS
============================================================================== */

func (c Link) HasRole(roles ...Role) bool {
	for _, existingRole := range c.Roles {
		for _, allowedRole := range roles {
			if existingRole == allowedRole {
				return true
			}
		}
	}
	return false
}

func (c *Link) AddRole(r Role) error {
	if r != Admin && r != Manager && sliceman.IndexOf(c.Application.PossibleRoles, r) == -1 {
		return ErrInvalidRole
	}

	if sliceman.IndexOf(c.Roles, r) != -1 {
		return ErrRedundantRequest
	}

	c.Roles = append(c.Roles, r)
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

func (c *Link) RemoveRole(r Role) error {
	if r != Admin && r != Manager && sliceman.IndexOf(c.Application.PossibleRoles, r) == -1 {
		return ErrInvalidRole
	}

	index := sliceman.IndexOf(c.Roles, r)
	if index == -1 {
		return ErrRedundantRequest
	}

	c.Roles = sliceman.Remove(c.Roles, index)
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

func (c *Link) RevokeConsent() error {
	c.HasConsent = false
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

func (c *Link) RestoreConsent() error {
	c.HasConsent = true
	c.UpdatedAt = time.Now()
	return validator.Validate(c)
}

/* ==============================================================================
	VIEWS
============================================================================== */

type LinkView struct {
	ID          uuid.UUID       `json:"id"`
	Application ApplicationView `json:"application"`
	Roles       []Role          `json:"roles"`
	HasConsent  bool            `json:"hasConsent"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (l Link) View() LinkView {
	app := l.Application.View()

	return LinkView{
		ID:          l.ID,
		Application: app,
		Roles:       l.Roles,
		HasConsent:  l.HasConsent,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}
