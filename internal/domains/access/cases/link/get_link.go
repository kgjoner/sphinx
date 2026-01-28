package linkcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetLink struct {
	AccessRepo access.Repo
}

type GetLinkInput struct {
	UserID        uuid.UUID    `json:"-"`
	ApplicationID uuid.UUID    `json:"-"`
	Actor         shared.Actor `json:"-"`
}

func (i GetLink) Execute(input GetLinkInput) (out access.LinkView, err error) {
	if err := access.CanReadLink(&input.Actor, input.UserID, input.ApplicationID); err != nil {
		return out, err
	}

	link, err := i.AccessRepo.GetUserLink(input.UserID, input.ApplicationID)
	if err != nil {
		return out, err
	} else if link == nil {
		return out, access.ErrLinkNotFound
	}

	return link.View(), nil
}
