package baserepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/cornucopia/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

func (q DAO) InsertAccount(acc *auth.Account) error {
	raw, exists := rawQueries["CreateAccount"]
	if !exists {
		return ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw,
		acc.Id,
		acc.Email.String(),
		acc.Password,
		datatransform.ToNullString(acc.Phone.String()),
		datatransform.ToNullString(acc.Username),
		datatransform.ToNullString(acc.Document.String()),
		datatransform.ToNullRawMessage(acc.ExtraData),
		acc.IsActive,
		datatransform.ToNullString(acc.PendingEmail.String()),
		datatransform.ToNullString(acc.PendingPhone.String()),
		acc.HasEmailBeenVerified,
		acc.HasPhoneBeenVerified,
		datatransform.ToRawMessage(acc.VerificationCodes),
	)
	err := row.Scan(&acc.InternalId)
	return err
}

func (q DAO) UpdateAccount(acc auth.Account) error {
	raw, exists := rawQueries["UpdateAccount"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.db.ExecContext(q.ctx, raw,
		acc.Id,
		acc.Email.String(),
		acc.Password,
		datatransform.ToNullString(acc.Phone.String()),
		datatransform.ToNullString(acc.Username),
		datatransform.ToNullString(acc.Document.String()),
		datatransform.ToNullRawMessage(acc.ExtraData),
		acc.IsActive,
		datatransform.ToNullString(acc.PendingEmail.String()),
		datatransform.ToNullString(acc.PendingPhone.String()),
		acc.HasEmailBeenVerified,
		acc.HasPhoneBeenVerified,
		datatransform.ToRawMessage(acc.VerificationCodes),
		acc.PasswordUpdatedAt,
		acc.UpdatedAt,
	)
	return err
}

func (q DAO) GetAccountById(id uuid.UUID) (*auth.Account, error) {
	raw, exists := rawQueries["GetAccountById"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, id)
	var item auth.Account
	err := row.Scan(
		&item.InternalId,
		&item.Id,
		&item.Email,
		&item.Phone,
		&item.Password,
		&item.Username,
		&item.Document,
		dbhandler.Struct(&item.ExtraData),
		&item.IsActive,
		&item.PendingEmail,
		&item.PendingPhone,
		&item.HasEmailBeenVerified,
		&item.HasPhoneBeenVerified,
		dbhandler.Map(&item.VerificationCodes),
		dbhandler.StructArray(&item.Links),
		dbhandler.StructArray(&item.ActiveSessions),
		&item.PasswordUpdatedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &item, nil
}

func (q DAO) GetAccountByEntry(entry auth.Entry) (*auth.Account, error) {
	raw, exists := rawQueries["GetAccountByEntry"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, entry.String())
	var item auth.Account
	err := row.Scan(
		&item.InternalId,
		&item.Id,
		&item.Email,
		&item.Phone,
		&item.Password,
		&item.Username,
		&item.Document,
		dbhandler.Struct(&item.ExtraData),
		&item.IsActive,
		&item.PendingEmail,
		&item.PendingPhone,
		&item.HasEmailBeenVerified,
		&item.HasPhoneBeenVerified,
		dbhandler.Map(&item.VerificationCodes),
		dbhandler.StructArray(&item.Links),
		dbhandler.StructArray(&item.ActiveSessions),
		&item.PasswordUpdatedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &item, nil
}

func (q DAO) GetAccountByLink(linkId uuid.UUID) (*auth.Account, error) {
	raw, exists := rawQueries["GetAccountByLink"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, linkId)
	var item auth.Account
	err := row.Scan(
		&item.InternalId,
		&item.Id,
		&item.Email,
		&item.Phone,
		&item.Password,
		&item.Username,
		&item.Document,
		dbhandler.Struct(&item.ExtraData),
		&item.IsActive,
		&item.PendingEmail,
		&item.PendingPhone,
		&item.HasEmailBeenVerified,
		&item.HasPhoneBeenVerified,
		dbhandler.Map(&item.VerificationCodes),
		dbhandler.StructArray(&item.Links),
		dbhandler.StructArray(&item.ActiveSessions),
		&item.PasswordUpdatedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &item, nil
}
