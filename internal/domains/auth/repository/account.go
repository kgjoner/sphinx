package authrepo

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/cornucopia/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	psqlrepo "github.com/kgjoner/sphinx/postgres"
)

func (r AuthRepo) InsertAccount(acc auth.Account) (int, error) {
	return r.q.CreateAccount(r.ctx, psqlrepo.CreateAccountParams{
		ID:                   acc.Id,
		Email:                acc.Email.String(),
		Password:             acc.Password,
		Phone:                datatransform.ToNullString(acc.Phone.String()),
		Username:             datatransform.ToNullString(acc.Username),
		Document:             datatransform.ToNullString(acc.Document.String()),
		IsActive:             acc.IsActive,
		HasEmailBeenVerified: acc.HasEmailBeenVerified,
		HasPhoneBeenVerified: acc.HasPhoneBeenVerified,
		Codes:                datatransform.ToRawMessage(acc.Codes),
	})
}

func (r AuthRepo) UpdateAccount(acc auth.Account) error {
	return r.q.UpdateAccount(r.ctx, psqlrepo.UpdateAccountParams{
		ID:                   acc.Id,
		Email:                acc.Email.String(),
		Password:             acc.Password,
		Phone:                datatransform.ToNullString(acc.Phone.String()),
		Username:             datatransform.ToNullString(acc.Username),
		Document:             datatransform.ToNullString(acc.Document.String()),
		IsActive:             acc.IsActive,
		HasEmailBeenVerified: acc.HasEmailBeenVerified,
		HasPhoneBeenVerified: acc.HasPhoneBeenVerified,
		Codes:                datatransform.ToRawMessage(acc.Codes),
		PasswordUpdatedAt:    datatransform.ToNullTime(acc.PasswordUpdatedAt),
		UpdatedAt:            acc.UpdatedAt,
	})
}

func (r AuthRepo) GetAccountById(id uuid.UUID) (*auth.Account, error) {
	return dbhandler.HandleSingleQuery[auth.Account](
		r.q.GetAccountById(r.ctx, id),
	)
}

func (r AuthRepo) GetAccountByEntry(entry string) (*auth.Account, error) {
	return dbhandler.HandleSingleQuery[auth.Account](
		r.q.GetAccountByEntry(r.ctx, entry),
	)
}
