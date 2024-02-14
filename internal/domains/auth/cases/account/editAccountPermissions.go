package accountcase

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type EditAccountPermissions struct {
	AuthRepo authcase.AuthRepo
}

type EditAccountPermissionsInput struct {
	TargetAccountEntry string
	Roles              []auth.Role
	Grantings         []string
	ShouldRemove       sql.NullBool
	Actor              auth.Account
}

func (i EditAccountPermissions) Execute(input EditAccountPermissionsInput) (*auth.AccountPrivateView, error) {
	if !input.ShouldRemove.Valid {
		return nil, normalizederr.NewRequestError("Must inform whether permissions should be added or removed.", "")
	}

	var targetAcc *auth.Account
	var err error
	if id, err := uuid.Parse(input.TargetAccountEntry); err != nil {
		targetAcc, err = i.AuthRepo.GetAccountById(id)
	} else {
		targetAcc, err = i.AuthRepo.GetAccountByEntry(input.TargetAccountEntry)
	}
	
	if err != nil {
	 return nil, err
	} else if targetAcc == nil {
	 return nil, normalizederr.NewRequestError("Account does not exist", "")
	}

	if input.Roles != nil {
		for _, r := range input.Roles {
			if input.ShouldRemove.Bool {
				err = targetAcc.RemoveRole(r, input.Actor)
			} else {
				err = targetAcc.AddRole(r, input.Actor)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	if input.Grantings != nil {
		for _, g := range input.Grantings {
			if input.ShouldRemove.Bool {
				err = targetAcc.RemoveGranting(g, input.Actor)
			} else {
				err = targetAcc.AddGranting(g, input.Actor)
			}
			if err != nil {
				return nil, err
			}
		}
	}

	err = i.AuthRepo.UpsertLinks(targetAcc.LinksToPersist()...)
	if err != nil {
	 return nil, err
	}

	return targetAcc.PrivateView(input.Actor)
}
