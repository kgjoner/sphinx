package identity

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/validator"
)

type ExternalCredential struct {
	UserID            uuid.UUID `validate:"required"`
	ProviderName      string    `validate:"required"`
	ProviderSubjectID string    `validate:"required"`
	ProviderAlias     string
	LastUsedAt        time.Time
	CreatedAt         time.Time `validate:"required"`
	UpdatedAt         time.Time `validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type ExternalCredentialCreationFields struct {
	ProviderName      string
	ProviderSubjectID string
	ProviderAlias     string
}

func (u *User) newExternalCredential(fields *ExternalCredentialCreationFields) (*ExternalCredential, error) {
	now := time.Now()
	created := &ExternalCredential{
		UserID:            u.ID,
		ProviderName:      fields.ProviderName,
		ProviderSubjectID: fields.ProviderSubjectID,
		ProviderAlias:     fields.ProviderAlias,
		LastUsedAt:        now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return created, validator.Validate(created)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (ec *ExternalCredential) MarkUse() error {
	ec.LastUsedAt = time.Now()
	ec.UpdatedAt = time.Now()
	return validator.Validate(ec)
}

func (ec *ExternalCredential) SetAlias(alias string) error {
	ec.ProviderAlias = alias
	ec.UpdatedAt = time.Now()
	return validator.Validate(ec)
}

/* ==============================================================================
	VIEWS
============================================================================== */

type ExternalCredentialView struct {
	UserID            uuid.UUID `json:"userId"`
	ProviderName      string    `json:"providerName"`
	ProviderSubjectID string    `json:"providerSubjectId"`
	ProviderAlias     string    `json:"providerAlias"`
	LastUsedAt        time.Time `json:"lastUsedAt"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (ec ExternalCredential) View() ExternalCredentialView {
	return ExternalCredentialView{
		UserID:            ec.UserID,
		ProviderName:      ec.ProviderName,
		ProviderSubjectID: ec.ProviderSubjectID,
		ProviderAlias:     ec.ProviderAlias,
		LastUsedAt:        ec.LastUsedAt,
		CreatedAt:         ec.CreatedAt,
		UpdatedAt:         ec.UpdatedAt,
	}
}
