package baserepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/utils/datatransform"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

func (q DAO) InsertUser(user *auth.User) error {
	raw, exists := rawQueries["CreateUser"]
	if !exists {
		return ErrNoQuery
	}

	row := q.db.QueryRowContext(q.ctx, raw,
		user.ID,
		user.Email.String(),
		user.Password,
		datatransform.ToNullString(user.Phone.String()),
		datatransform.ToNullString(user.Username),
		datatransform.ToNullString(user.Document.String()),
		datatransform.ToNullRawMessage(user.ExtraData),
		user.IsActive,
		datatransform.ToNullString(user.PendingEmail.String()),
		datatransform.ToNullString(user.PendingPhone.String()),
		user.HasEmailBeenVerified,
		user.HasPhoneBeenVerified,
		datatransform.ToRawMessage(user.VerificationCodes),
		datatransform.ToRawMessage(user.ExternalAuthIDs),
	)
	err := row.Scan(&user.InternalID)
	return err
}

func (q DAO) UpdateUser(user auth.User) error {
	raw, exists := rawQueries["UpdateUser"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.db.ExecContext(q.ctx, raw,
		user.ID,
		user.Email.String(),
		user.Password,
		datatransform.ToNullString(user.Phone.String()),
		datatransform.ToNullString(user.Username),
		datatransform.ToNullString(user.Document.String()),
		datatransform.ToNullRawMessage(user.ExtraData),
		user.IsActive,
		datatransform.ToNullString(user.PendingEmail.String()),
		datatransform.ToNullString(user.PendingPhone.String()),
		user.HasEmailBeenVerified,
		user.HasPhoneBeenVerified,
		datatransform.ToRawMessage(user.VerificationCodes),
		datatransform.ToRawMessage(user.ExternalAuthIDs),
		user.PasswordUpdatedAt,
		user.UpdatedAt,
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
