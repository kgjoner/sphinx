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
	Id          uuid.UUID   `json:"id" validator:"required"`
	AccountId   int         `json:"-" validator:"required"`
	Application Application `json:"application" validator:"required"`
	Roles       []Role      `json:"roles"`
	Grantings   []string    `json:"grantings"`

	CreatedAt time.Time `json:"createdAt" validator:"required"`
	UpdatedAt time.Time `json:"updatedAt" validator:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

func (a *Account) CreateLinkTo(app Application) (*Link, error) {
	for _, l := range a.Links {
		if l.Application.Id == app.Id {
			return nil, normalizederr.NewRequestError("Account has already been linked to desired application.", "")
		}
	}

	now := time.Now()
	created := &Link{
		Id:          uuid.New(),
		AccountId:   a.InternalId,
		Application: app,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, validator.Validate(created)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (l *Link) AddRole(r Role, actor Account) error {
	if !actor.IsAdminOn(l.Application) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	if sliceman.IndexOf(l.Roles, r) != -1 {
		return normalizederr.NewRequestError("Role has already been added.", "")
	}

	l.Roles = append(l.Roles, r)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) RemoveRole(r Role, actor Account) error {
	if !actor.IsAdminOn(l.Application) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	index := sliceman.IndexOf(l.Roles, r)
	if index == -1 {
		return normalizederr.NewRequestError("Role has not been added.", "")
	}

	l.Roles = sliceman.Remove(l.Roles, index)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) AddGranting(g string, actor Account) error {
	if !actor.IsAdminOn(l.Application) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	if sliceman.IndexOf(l.Application.Grantings, g) == -1 {
		return normalizederr.NewForbiddenError("Application does not support the desired granting.")
	}

	if sliceman.IndexOf(l.Grantings, g) != -1 {
		return normalizederr.NewRequestError("Granting has already been added.", "")
	}

	l.Grantings = append(l.Grantings, g)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}

func (l *Link) RemoveGranting(g string, actor Account) error {
	if !actor.IsAdminOn(l.Application) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	if sliceman.IndexOf(l.Application.Grantings, g) == -1 {
		return normalizederr.NewForbiddenError("Application does not support the desired granting.")
	}

	index := sliceman.IndexOf(l.Grantings, g)
	if index == -1 {
		return normalizederr.NewRequestError("Granting has not been added.", "")
	}

	l.Grantings = sliceman.Remove(l.Grantings, index)
	l.UpdatedAt = time.Now()
	return validator.Validate(l)
}
