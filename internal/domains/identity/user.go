package identity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/cornucopia/v2/utils/pwdgen"
	"github.com/kgjoner/cornucopia/v2/utils/structop"
	"github.com/kgjoner/sphinx/internal/shared"
)

type User struct {
	InternalID int
	ID         uuid.UUID    `validate:"required"`
	Email      htypes.Email `validate:"required"`
	Phone      htypes.PhoneNumber
	Password   shared.HashedPassword `validate:"required"`
	Username   string                `validate:"wordID,atLeastOne=letter"`
	Document   htypes.Document
	ExtraData

	PendingEmail         htypes.Email
	HasEmailBeenVerified bool
	PendingPhone         htypes.PhoneNumber
	HasPhoneBeenVerified bool
	VerificationCodes    map[VerificationKind]string
	PasswordUpdatedAt    htypes.NullTime
	UsernameUpdatedAt    htypes.NullTime

	ExternalCredentials []ExternalCredential
	IsActive            bool
	CreatedAt           time.Time `validate:"required"`
	UpdatedAt           time.Time `validate:"required"`
}

type ExtraData struct {
	Name    string         `json:"name"`
	Surname string         `json:"surname"`
	Address htypes.Address `json:"address"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type UserCreationFields struct {
	Email    htypes.Email          `validate:"required"`
	Password shared.HashedPassword `json:"-" validate:"required"`
	Phone    htypes.PhoneNumber
	Username string
	Document htypes.Document
	UserExtraFields
}

func NewUser(f *UserCreationFields) (*User, error) {
	if f.Password.IsZero() {
		return nil, shared.ErrEmptyPassword
	}

	now := time.Now()
	user := &User{
		ID:       uuid.New(),
		Email:    f.Email,
		Phone:    f.Phone,
		Password: f.Password,
		Username: strings.ToLower(f.Username),
		Document: f.Document,
		ExtraData: ExtraData{
			Name:    f.Name,
			Surname: f.Surname,
			Address: f.Address,
		},

		IsActive:          true,
		VerificationCodes: map[VerificationKind]string{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	user.generateCodeFor(VerificationEmail)
	if !user.Phone.IsZero() {
		user.generateCodeFor(VerificationPhone)
	}

	return user, validator.Validate(user)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (u User) Name() string {
	if u.ExtraData.Name != "" {
		return u.ExtraData.Name
	}

	if u.Username != "" {
		return u.Username
	}

	email := u.Email.String()
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at]
	}
	return email
}

func (u *User) VerifyUser(kind VerificationKind, code string) error {
	err := validator.Validate(code, "required")
	if err != nil {
		return err
	}

	err = validator.Validate(kind)
	if err != nil {
		return err
	}

	switch kind {
	case VerificationEmail:
		if u.HasEmailBeenVerified && u.PendingEmail.IsZero() {
			return apperr.NewRequestError("Email has already been verified.")
		}
	case VerificationPhone:
		if u.HasPhoneBeenVerified && u.PendingPhone.IsZero() {
			return apperr.NewRequestError("Phone has already been verified.")
		}
	}

	if u.VerificationCodes[kind] == "" {
		return apperr.NewConflictError("User does not have a verification code.")
	}

	if code != u.VerificationCodes[kind] {
		return apperr.NewRequestError("Invalid code.")
	}

	switch kind {
	case VerificationEmail:
		if !u.PendingEmail.IsZero() {
			// Confirm email update
			u.Email = u.PendingEmail
			var emptyEmail htypes.Email
			u.PendingEmail = emptyEmail
		}
		u.HasEmailBeenVerified = true
	case VerificationPhone:
		if !u.PendingPhone.IsZero() {
			// Confirm phone update
			u.Phone = u.PendingPhone
			var emptyPhone htypes.PhoneNumber
			u.PendingPhone = emptyPhone
		}
		u.HasPhoneBeenVerified = true
	}
	u.clearCodeFor(kind)
	u.UpdatedAt = time.Now()

	return nil
}

func (u *User) ChangePassword(proof shared.AuthProof, newPw shared.HashedPassword) error {
	switch proofTyped := proof.(type) {
	case *shared.PasswordProof:
		if !proofTyped.ValidFor(u.Password) {
			return shared.ErrInvalidCredentials
		}
	case *shared.CodeProof:
		if !proofTyped.ValidFor(u.VerificationCodes[VerificationPasswordReset]) {
			return shared.ErrInvalidCode
		}
		u.clearCodeFor(VerificationPasswordReset)
	default:
		return shared.ErrInvalidProof
	}

	now := time.Now()
	u.Password = newPw
	u.PasswordUpdatedAt = htypes.NullTime{Time: now}
	u.UpdatedAt = now
	return validator.Validate(u)
}

func (u *User) RequestPasswordReset() (string, error) {
	u.generateCodeFor(VerificationPasswordReset)
	u.UpdatedAt = time.Now()
	return u.VerificationCodes[VerificationPasswordReset], validator.Validate(u)
}

func (u *User) UpdateEmail(email htypes.Email) error {
	if email.IsZero() {
		return ErrEmptyInput
	}

	if email == u.Email {
		return ErrRedundantRequest
	}

	u.PendingEmail = email
	u.generateCodeFor(VerificationEmail)
	u.UpdatedAt = time.Now()
	return validator.Validate(u)
}

func (u *User) UpdateUsername(username string) error {
	if username == "" {
		return ErrEmptyInput
	}

	if username == u.Username {
		return ErrRedundantRequest
	}

	now := time.Now()
	u.Username = strings.ToLower(username)
	u.UsernameUpdatedAt = htypes.NullTime{Time: now}
	u.UpdatedAt = now
	return validator.Validate(u)
}

func (u *User) UpdatePhone(phone htypes.PhoneNumber) error {
	if phone.IsZero() {
		return ErrEmptyInput
	}

	if phone == u.Phone {
		return ErrRedundantRequest
	}

	u.PendingPhone = phone
	u.generateCodeFor(VerificationPhone)
	u.UpdatedAt = time.Now()
	return validator.Validate(u)
}

func (u *User) UpdateDocument(document htypes.Document) error {
	if document.IsZero() {
		return ErrEmptyInput
	}

	if document == u.Document {
		return ErrRedundantRequest
	}

	u.Document = document
	u.UpdatedAt = time.Now()
	return validator.Validate(u)
}

type UserExtraFields struct {
	Name    string         `json:"name"`
	Surname string         `json:"surname"`
	Address htypes.Address `json:"address"`
}

func (u *User) UpdateExtraData(f UserExtraFields) error {
	structop.New(&u.ExtraData).Update(f)
	u.UpdatedAt = time.Now()
	return validator.Validate(u)
}

// Allows users to cancel their pending field update (email or phone).
// It will remove the pending field and clear the code for it.
func (u *User) CancelPendingField(field string) error {
	err := validator.Validate(field, "required", "oneof=email phone")
	if err != nil {
		return err
	}

	switch field {
	case "email":
		return u.cancelPendingEmail()
	case "phone":
		return u.cancelPendingPhone()
	default:
		return apperr.NewRequestError("Invalid field name.")
	}
}

// Allows users to cancel their pending email update
func (u *User) cancelPendingEmail() error {
	if u.PendingEmail.IsZero() {
		return apperr.NewRequestError("No pending email update to cancel.")
	}

	var emptyEmail htypes.Email
	u.PendingEmail = emptyEmail
	u.clearCodeFor(VerificationEmail)
	u.UpdatedAt = time.Now()

	return nil
}

// Allows users to cancel their pending phone update
func (u *User) cancelPendingPhone() error {
	if u.PendingPhone.IsZero() {
		return apperr.NewRequestError("No pending phone update to cancel.")
	}

	var emptyPhone htypes.PhoneNumber
	u.PendingPhone = emptyPhone
	u.clearCodeFor(VerificationPhone)
	u.UpdatedAt = time.Now()

	return nil
}

// It will generate an random string and save it in the field Code, under the desired key (kind).
// It overwrites old value, if any.
func (u *User) generateCodeFor(kind VerificationKind) {
	u.VerificationCodes[kind] = pwdgen.GeneratePassword(12, "lower", "upper", "number")
}

func (u *User) clearCodeFor(kind VerificationKind) {
	delete(u.VerificationCodes, kind)
}

func (u *User) AddExternalCredential(f *ExternalCredentialCreationFields) (*ExternalCredential, error) {
	ec, err := u.newExternalCredential(f)
	if err != nil {
		return nil, err
	}

	if u.ExternalCredentials == nil {
		u.ExternalCredentials = []ExternalCredential{}
	}

	u.ExternalCredentials = append(u.ExternalCredentials, *ec)
	u.UpdatedAt = time.Now()
	return ec, nil
}

func (u *User) ExternalCredential(providerName string, providerSubjectID string) *ExternalCredential {
	for _, ec := range u.ExternalCredentials {
		if ec.ProviderName == providerName && ec.ProviderSubjectID == providerSubjectID {
			return &ec
		}
	}

	return nil
}

/* ==============================================================================
	VIEWS
============================================================================== */

type UserView struct {
	ID       uuid.UUID          `json:"id" validate:"required"`
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty"`
	Document htypes.Document    `json:"document,omitempty"`
	Name     string             `json:"name,omitempty"`
	Surname  string             `json:"surname,omitempty"`
	Address  *htypes.Address    `json:"address,omitempty"`

	PendingEmail         htypes.Email       `json:"pendingEmail,omitempty"`
	HasEmailBeenVerified bool               `json:"hasEmailBeenVerified"`
	PendingPhone         htypes.PhoneNumber `json:"pendingPhone,omitempty"`
	HasPhoneBeenVerified bool               `json:"hasPhoneBeenVerified"`
	UsernameUpdatedAt    htypes.NullTime    `json:"usernameUpdatedAt"`

	ExternalCredentials []ExternalCredentialView `json:"externalCredentials,omitempty"`
	IsActive            bool                     `json:"isActive"`
}

func (u User) View() UserView {
	var address *htypes.Address
	if u.ExtraData.Address != (htypes.Address{}) {
		address = &u.ExtraData.Address
	}

	extCredentials := []ExternalCredentialView{}
	for _, ec := range u.ExternalCredentials {
		extCredentials = append(extCredentials, ec.View())
	}

	return UserView{
		u.ID,
		u.Email,
		u.Phone,
		u.Username,
		u.Document,
		u.ExtraData.Name,
		u.ExtraData.Surname,
		address,
		u.PendingEmail,
		u.HasEmailBeenVerified,
		u.PendingPhone,
		u.HasPhoneBeenVerified,
		u.UsernameUpdatedAt,
		extCredentials,
		u.IsActive,
	}
}

type UserLeanView struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	Username string    `json:"username,omitempty"`
	Name     string    `json:"name,omitempty"`
	Surname  string    `json:"surname,omitempty"`

	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

func (u User) LeanView() UserLeanView {
	return UserLeanView{
		u.ID,
		u.Username,
		u.ExtraData.Name,
		u.ExtraData.Surname,
		u.IsActive,
		u.CreatedAt,
		u.UpdatedAt,
	}
}
