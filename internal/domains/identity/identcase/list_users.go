package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ListUsers struct {
	IdentityRepo identity.Repo
}

type ListUsersInput struct {
	SearchFilter string            `json:"-"`
	View         string            `json:"-" validate:"oneof=lean full"`
	Pagination   prim.Pagination `json:"-"`
	Actor        shared.Actor      `json:"-"`
}

func (u ListUsers) Execute(input ListUsersInput) (out *prim.PaginatedData[any], err error) {
	if err := identity.CanListUsers(&input.Actor, input.SearchFilter); err != nil {
		return out, err
	}

	users, err := u.IdentityRepo.ListUsers(input.SearchFilter, &input.Pagination)
	if err != nil {
		return out, err
	}

	if input.View == "full" {
		if err := identity.CanReadUserSensitiveData(&input.Actor, uuid.Nil); err != nil {
			return out, err
		}

		return prim.TransformPaginatedData(users, func(data identity.User) any {
			return data.View()
		}), nil
	}

	return prim.TransformPaginatedData(users, func(data identity.User) any {
		return data.LeanView()
	}), nil

}
