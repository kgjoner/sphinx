package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/sliceman"
	"github.com/kgjoner/cornucopia/utils/structop"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	InternalId int                `json:"-"`
	Id         uuid.UUID          `json:"id" validator:"required"`
	Email      htypes.Email       `json:"email" validator:"required"`
	Phone      htypes.PhoneNumber `json:"phone"`
	Password   string             `json:"password"`
	Username   string             `json:"username" validator:"wordId"`
	Document   htypes.Document    `json:"document"`

	IsActive               bool                       `json:"isActive"`
	HasEmailBeenVerified   bool                       `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified   bool                       `json:"hasPhoneBeenVerified"`
	Codes                  map[AccountCodeKind]string `json:"-"`
	Links                  []Link                     `json:"-"`
	ActiveSessions         []Session                  `json:"-"`
	justTerminatedSessions []Session                  `json:"-"`

	PasswordResetAt time.Time `json:"-"`
	CreatedAt       time.Time `json:"createdAt" validator:"required"`
	UpdatedAt       time.Time `json:"updatedAt" validator:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type AccountCreationFields struct {
	Email    htypes.Email       `json:"email"`
	Phone    htypes.PhoneNumber `json:"phone"`
	Password string             `json:"password"`
	Username string             `json:"username"`
	Document htypes.Document    `json:"document"`
}

func NewAccount(a *AccountCreationFields) (*Account, error) {
	now := time.Now()
	acc := &Account{
		Id:       uuid.New(),
		Email:    a.Email,
		Phone:    a.Phone,
		Password: hashData(a.Password),
		Username: a.Username,
		Document: a.Document,

		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	acc.GenerateCodeFor(AccountCodeKindValues.EMAIL_VERIFICATION)
	if acc.Phone != "" {
		acc.GenerateCodeFor(AccountCodeKindValues.PHONE_VERIFICATION)
	}

	return acc, validator.Validate(acc)
}

func hashData(str string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(str), 14)
	return string(bytes)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (a Account) DoesPasswordMatch(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

type AccountMissableFields struct {
	Phone    htypes.PhoneNumber `json:"phone"`
	Username string             `json:"username"`
	Document htypes.Document    `json:"document"`
}

func (a *Account) AddMissingFields(f AccountMissableFields) error {
	if !f.Phone.IsZero() && !a.Phone.IsZero() {
		return normalizederr.NewRequestError("Phone was already added.", "")
	}
	if f.Username != "" && a.Username != "" {
		return normalizederr.NewRequestError("Username was already added.", "")
	}
	if !f.Document.IsZero() && !a.Document.IsZero() {
		return normalizederr.NewRequestError("Document was already added.", "")
	}

	structop.New(a).Update(f)
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

// It will generate an uuid and save it in the field Code, under the desired key (kind).
// It overwrites old value, if any.
func (a *Account) GenerateCodeFor(kind AccountCodeKind) {
	a.Codes[kind] = uuid.New().String()
}

func (a *Account) ClearCodeFor(kind AccountCodeKind) {
	delete(a.Codes, kind)
}

func (a *Account) ResetPassword(newPassword string, code string) error {
	resetCode := a.Codes[AccountCodeKindValues.PASSWORD_RESET]
	if resetCode == "" {
		return normalizederr.NewRequestError("Account did not request a password reset.", "")
	} else if code != a.Codes[AccountCodeKindValues.PASSWORD_RESET] {
		return normalizederr.NewForbiddenError("Invalid code.")
	}

	a.Password = newPassword
	a.ClearCodeFor(AccountCodeKindValues.PASSWORD_RESET)

	now := time.Now()
	a.PasswordResetAt = now
	a.UpdatedAt = now
	return validator.Validate(a)
}

/* ==============================================================================
	LINK RELATED METHODS
============================================================================== */

// Return link to desired application or nil if account is not linked to it.
func (a *Account) Link(app Application) *Link {
	for _, l := range a.Links {
		if l.Application.Id == app.Id {
			return &l
		}
	}

	return nil
}

func (a *Account) IsAdminOn(app Application) bool {
	link := a.Link(app)
	if link == nil {
		return false
	}

	for _, r := range link.Roles {
		if r == RoleValues.ADMIN {
			return true
		}
	}

	return false
}

/* ==============================================================================
	SESSION RELATED METHODS
============================================================================== */

// Return desired active session or nil if it does not exist
func (a *Account) Session(sessionId uuid.UUID) *Session {
	for _, s := range a.ActiveSessions {
		if s.Id == sessionId && s.IsActive {
			return &s
		}
	}

	return nil
}

// Return active sessions of the desired application
func (a *Account) Sessions(app Application) []Session {
	sessions := []Session{}
	for _, s := range a.ActiveSessions {
		if s.Application.Id == app.Id && s.IsActive {
			sessions = append(sessions, s)
		}
	}

	return sessions
}

// Return sessions that have just been updated
func (a *Account) SessionsToPersist() []Session {
	sessions := []Session{}
	for _, s := range a.ActiveSessions {
		if s.UpdatedAt.After(time.Now().Add(time.Duration(-5) * time.Second)) {
			sessions = append(sessions, s)
		}
	}

	return sessions
}

func (a *Account) TerminateSession(sessionId uuid.UUID) (*Session, error) {
	for i, s := range a.ActiveSessions {
		if s.Id == sessionId {
			err := s.terminate()
			if err != nil {
				return nil, err
			}

			if a.justTerminatedSessions == nil {
				a.justTerminatedSessions = []Session{}
			}
			a.justTerminatedSessions = append(a.justTerminatedSessions, s)
			sliceman.Remove(a.ActiveSessions, i)

			return &s, nil
		}
	}

	return nil, normalizederr.NewRequestError("Session not found.", "")
}

func (a *Account) TerminateAllSessions() error {
	if a.justTerminatedSessions == nil {
		a.justTerminatedSessions = []Session{}
	}

	for _, s := range a.ActiveSessions {
		err := s.terminate()
		if err != nil {
			return err
		}

		a.justTerminatedSessions = append(a.justTerminatedSessions, s)
	}

	a.ActiveSessions = []Session{}
	return nil
}

// Check if access token is valid and return its related session
func (a *Account) VerifyAccessToken(token *Jwt) (*Session, error) {
	if token.Claims.Kind != "access" {
		return nil, normalizederr.NewRequestError("Non access token.", "")
	}
	
	return a.verifyToken(token)
}

// Check if refresh token is valid and return its related session
func (a *Account) VerifyRefreshToken(token *Jwt) (*Session, error) {
	if token.Claims.Kind != "refresh" {
		return nil, normalizederr.NewRequestError("Non refresh token.", "")
	}

	s, err := a.verifyToken(token)
	if err != nil {
		return nil, err
	}

	doesMatch := s.doesRefreshTokenMatch(token.String())
	if !doesMatch {
		return nil, normalizederr.NewUnauthorizedError("Revoked token.")
	}

	return s, nil
}

func (a *Account) verifyToken(token *Jwt) (*Session, error) {
	if token.IsExpired() {
		return nil, normalizederr.NewUnauthorizedError("Expired token")
	}

	s := a.Session(token.Claims.SessionId)
	if s == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid session")
	}

	if token.Claims.Sub != a.Id || token.Claims.Aud != s.Application.Id {
		return nil, normalizederr.NewUnauthorizedError("Mismatched authentication")
	}

	return s, nil
}

func (a *Account) IssueNewTokens(refreshToken *Jwt) (*Jwt, *Jwt, error) {
	s, err := a.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, err
	}

	newRefreshToken, err := newJwt(*a, s.Id, "refresh")
	if err != nil {
		return nil, nil, err
	}

	s.updateRefreshToken(*newRefreshToken)

	newAccessToken, err := newJwt(*a, s.Id)
	if err != nil {
		return nil, nil, err
	}

	return newAccessToken, newRefreshToken, nil
}
