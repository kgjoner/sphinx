package accesscase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/shared"
)

type RemoveRole struct {
	AccessRepo access.Repo
}

type RemoveRoleInput struct {
	UserID        uuid.UUID    `json:"-"`
	ApplicationID uuid.UUID    `json:"-"`
	Role          access.Role  `validate:"required"`
	Actor         shared.Actor `json:"-"`
}

func (i RemoveRole) Execute(input RemoveRoleInput) (out bool, err error) {
	if input.Role == "" {
		return out, access.ErrEmptyRole
	}

	if err := access.CanManageRole(&input.Actor, input.ApplicationID, input.Role); err != nil {
		return out, err
	}

	link, err := i.AccessRepo.GetUserLink(input.UserID, input.ApplicationID)
	if err != nil {
		return out, err
	} else if link == nil {
		return out, access.ErrLinkNotFound
	}

	err = link.RemoveRole(input.Role)
	if err != nil {
		return out, err
	}

	err = i.AccessRepo.UpsertLinks(*link)
	if err != nil {
		return out, err
	}

	return true, nil
}
