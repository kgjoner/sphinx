package baserepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/utils/datatransform"
	"github.com/kgjoner/cornucopia/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

func (q DAO) InsertUser(acc *auth.User) error {
	raw, exists := rawQueries["CreateUser"]
	if !exists {
		return ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw,
		acc.ID,
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
		datatransform.ToRawMessage(acc.ExternalAuthIDs),
	)
	err := row.Scan(&acc.InternalID)
	return err
}

func (q DAO) UpdateUser(acc auth.User) error {
	raw, exists := rawQueries["UpdateUser"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.db.ExecContext(q.ctx, raw,
		acc.ID,
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
		datatransform.ToRawMessage(acc.ExternalAuthIDs),
		acc.PasswordUpdatedAt,
		acc.UpdatedAt,
	)
	return err
}

func (q DAO) GetUserByID(id uuid.UUID) (*auth.User, error) {
	raw, exists := rawQueries["GetUserByID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, id)
	var item auth.User
	err := row.Scan(
		&item.InternalID,
		&item.ID,
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
		dbhandler.Map(&item.ExternalAuthIDs),
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

func (q DAO) GetUserByEntry(entry auth.Entry) (*auth.User, error) {
	raw, exists := rawQueries["GetUserByEntry"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, entry.String())
	var item auth.User
	err := row.Scan(
		&item.InternalID,
		&item.ID,
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
		dbhandler.Map(&item.ExternalAuthIDs),
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

func (q DAO) GetUserByLink(linkID uuid.UUID) (*auth.User, error) {
	raw, exists := rawQueries["GetUserByLink"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, linkID)
	var item auth.User
	err := row.Scan(
		&item.InternalID,
		&item.ID,
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
		dbhandler.Map(&item.ExternalAuthIDs),
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

func (q DAO) GetUserByExternalAuthID(provider string, id string) (*auth.User, error) {
	raw, exists := rawQueries["GetUserByExternalAuthID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw, provider, id)
	var item auth.User
	err := row.Scan(
		&item.InternalID,
		&item.ID,
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
		dbhandler.Map(&item.ExternalAuthIDs),
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
