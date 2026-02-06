package identrepo

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/utils/datatransform"
	"github.com/kgjoner/cornucopia/v2/utils/dbhandler"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

func (q DAO) InsertUser(user *identity.User) error {
	raw, exists := rawQueries["CreateUser"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
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
	)
	return err
}

func (q DAO) UpdateUser(user identity.User) error {
	raw, exists := rawQueries["UpdateUser"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
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
		user.PasswordUpdatedAt,
		user.UpdatedAt,
	)
	return err
}

func (q DAO) GetUserByID(id uuid.UUID) (*identity.User, error) {
	raw, exists := rawQueries["GetUserByID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, id)
	var item identity.User
	err := row.Scan(
		&item.ID,
		&item.Email,
		&item.Phone,
		&item.Password,
		&item.Username,
		&item.Document,
		dbhandler.FromJSON(&item.ExtraData),
		&item.IsActive,
		&item.PendingEmail,
		&item.PendingPhone,
		&item.HasEmailBeenVerified,
		&item.HasPhoneBeenVerified,
		dbhandler.Map(&item.VerificationCodes),
		&item.PasswordUpdatedAt,
		dbhandler.FromJSON(&item.ExternalCredentials),
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

func (q DAO) GetUserByEntry(entry shared.Entry) (*identity.User, error) {
	raw, exists := rawQueries["GetUserByEntry"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, entry.String())
	var item identity.User
	err := row.Scan(
		&item.ID,
		&item.Email,
		&item.Phone,
		&item.Password,
		&item.Username,
		&item.Document,
		dbhandler.FromJSON(&item.ExtraData),
		&item.IsActive,
		&item.PendingEmail,
		&item.PendingPhone,
		&item.HasEmailBeenVerified,
		&item.HasPhoneBeenVerified,
		dbhandler.Map(&item.VerificationCodes),
		&item.PasswordUpdatedAt,
		dbhandler.FromJSON(&item.ExternalCredentials),
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

func (q DAO) GetUserByExternalCredential(provider string, subjectID string) (*identity.User, error) {
	raw, exists := rawQueries["GetUserByExternalCredential"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, provider, subjectID)
	var item identity.User
	err := row.Scan(
		&item.ID,
		&item.Email,
		&item.Phone,
		&item.Password,
		&item.Username,
		&item.Document,
		dbhandler.FromJSON(&item.ExtraData),
		&item.IsActive,
		&item.PendingEmail,
		&item.PendingPhone,
		&item.HasEmailBeenVerified,
		&item.HasPhoneBeenVerified,
		dbhandler.Map(&item.VerificationCodes),
		&item.PasswordUpdatedAt,
		dbhandler.FromJSON(&item.ExternalCredentials),
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

func (q DAO) ListUsers(filter string, pag *htypes.Pagination) (*htypes.PaginatedData[identity.User], error) {
	raw, exists := rawQueries["ListUsers"]
	if !exists {
		return nil, ErrNoQuery
	}

	rows, err := q.dbtx.QueryContext(q.ctx, raw, filter, pag.Limit+1, pag.Offset())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return handleListQuery(rows, pag, func(item *identity.User) []any {
		return []any{
			&item.ID,
			&item.Email,
			&item.Phone,
			&item.Password,
			&item.Username,
			&item.Document,
			dbhandler.FromJSON(&item.ExtraData),
			&item.IsActive,
			&item.PendingEmail,
			&item.PendingPhone,
			&item.HasEmailBeenVerified,
			&item.HasPhoneBeenVerified,
			dbhandler.Map(&item.VerificationCodes),
			&item.PasswordUpdatedAt,
			&item.UsernameUpdatedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		}
	})
}
