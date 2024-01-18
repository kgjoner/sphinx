package auth

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/sliceman"
	"github.com/kgjoner/cornucopia/utils/structop"
	"github.com/kgjoner/sphinx/internal/config"
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

	PasswordUpdatedAt time.Time `json:"-"`
	CreatedAt         time.Time `json:"createdAt" validator:"required"`
	UpdatedAt         time.Time `json:"updatedAt" validator:"required"`
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

	acc.generateCodeFor(AccountCodeKindValues.EMAIL_VERIFICATION)
	if !acc.Phone.IsZero() {
		acc.generateCodeFor(AccountCodeKindValues.PHONE_VERIFICATION)
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

func (a *Account) ChangePassword(oldPassword string, newPassword string) error {
	if !a.DoesPasswordMatch(oldPassword) {
		return normalizederr.NewForbiddenError("Invalid credentials.")
	}

	a.Password = newPassword

	now := time.Now()
	a.PasswordUpdatedAt = now
	a.UpdatedAt = now
	return validator.Validate(a)
}

func (a *Account) RequestPasswordReset() (string, error) {
	a.generateCodeFor(AccountCodeKindValues.PASSWORD_RESET)
	a.UpdatedAt = time.Now()
	return a.Codes[AccountCodeKindValues.PASSWORD_RESET], validator.Validate(a)
}

func (a *Account) ResetPassword(newPassword string, code string) error {
	resetCode := a.Codes[AccountCodeKindValues.PASSWORD_RESET]
	if resetCode == "" {
		return normalizederr.NewRequestError("Account did not request a password reset.", "")
	} else if code != a.Codes[AccountCodeKindValues.PASSWORD_RESET] {
		return normalizederr.NewForbiddenError("Invalid code.")
	}

	a.Password = newPassword
	a.clearCodeFor(AccountCodeKindValues.PASSWORD_RESET)

	now := time.Now()
	a.PasswordUpdatedAt = now
	a.UpdatedAt = now
	return validator.Validate(a)
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

	if !f.Phone.IsZero() {
		a.generateCodeFor(AccountCodeKindValues.PHONE_VERIFICATION)
	}
	return validator.Validate(a)
}

// It will generate an uuid and save it in the field Code, under the desired key (kind).
// It overwrites old value, if any.
func (a *Account) generateCodeFor(kind AccountCodeKind) {
	a.Codes[kind] = uuid.New().String()
}

func (a *Account) clearCodeFor(kind AccountCodeKind) {
	delete(a.Codes, kind)
}

/* ==============================================================================
	LINK RELATED METHODS
============================================================================== */

// Create link to desired application.
func (a *Account) LinkTo(app Application) error {
	for _, l := range a.Links {
		if l.Application.Id == app.Id {
			return normalizederr.NewRequestError("Account has already been linked to desired application.", "")
		}
	}

	link := newLink(a, app)
	a.Links = append(a.Links, *link)
	return validator.Validate(a)
}

func (a Account) HasRole(app Application, roles ...Role) bool {
	link := a.link(app)
	if link == nil {
		return false
	}

	return link.hasRole(roles...)
}

func (a *Account) AddRole(r Role, app Application, actor Account) error {
	return a.updatePermission(app, actor, func(l *Link) error {
		return l.addRole(r)
	})
}

func (a *Account) RemoveRole(r Role, app Application, actor Account) error {
	return a.updatePermission(app, actor, func(l *Link) error {
		return l.removeRole(r)
	})
}

func (a *Account) AddGranting(g string, app Application, actor Account) error {
	return a.updatePermission(app, actor, func(l *Link) error {
		return l.addGranting(g)
	})
}

func (a *Account) RemoveGranting(g string, app Application, actor Account) error {
	return a.updatePermission(app, actor, func(l *Link) error {
		return l.removeGranting(g)
	})
}

// Return links that have just been updated
func (a *Account) LinksToPersist() []Link {
	links := []Link{}
	for _, l := range a.Links {
		if l.UpdatedAt.After(time.Now().Add(time.Duration(-5) * time.Second)) {
			links = append(links, l)
		}
	}

	return links
}

// Prepare and execute desired role and/or granting updates
func (a *Account) updatePermission(app Application, actor Account, updaterFn func(*Link) error) error {
	if !actor.HasRole(app, RoleValues.ADMIN) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	link := a.link(app)
	if link == nil {
		return normalizederr.NewRequestError("Target account is not linked to desired application.", "")
	}

	err := updaterFn(link)
	if err != nil {
		return err
	}

	for i, l := range a.Links {
		if l.Id == link.Id {
			a.Links[i] = *link
		}
	}

	return nil
}

// Return desired application link or nil if account is not linked to it.
func (a *Account) link(app Application) *Link {
	for _, l := range a.Links {
		if l.Application.Id == app.Id {
			return &l
		}
	}

	return nil
}

/* ==============================================================================
	SESSION RELATED METHODS
============================================================================== */

// Create a new session and generate tokens
func (a *Account) InitSession(f *SessionCreationFields) (*authToken, *authToken, error) {
	// Check if link exists
	link := a.link(f.Application)
	if link == nil {
		return nil, nil, normalizederr.NewRequestError("Account is not linked to desired application.", "")
	}

	// Terminate exceeding sessions
	if concurrentSessions :=
		SessionSortableByAge(a.sessionsByApp(f.Application)); config.Environment.MAX_CONCURRENT_SESSIONS > 0 &&
		len(concurrentSessions) >= config.Environment.MAX_CONCURRENT_SESSIONS {

		previousExcess := len(concurrentSessions) - config.Environment.MAX_CONCURRENT_SESSIONS
		/* If all goes well, previousExcess will be zero. There will be only one session to terminate
		for preventing new session to exceed max limit */
		sort.Sort(concurrentSessions)
		sessionsToTerminate := concurrentSessions[:(previousExcess + 1)]
		for _, s := range sessionsToTerminate {
			_, err := a.TerminateSession(s.Id)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	// Create session and tokens
	session := newSession(*a, f)

	refreshToken, err := newAuthToken(authTokenCreationFields{*a, session.Id, true})
	if err != nil {
		return nil, nil, err
	}

	session.updateRefreshToken(*refreshToken)

	accessToken, err := newAuthToken(authTokenCreationFields{*a, session.Id, false})
	if err != nil {
		return nil, nil, err
	}

	a.ActiveSessions = append(a.ActiveSessions, *session)
	return accessToken, refreshToken, validator.Validate(a)
}

// Generate new tokens for an existing session
func (a *Account) IssueNewTokens(refreshToken *authToken) (*authToken, *authToken, error) {
	s, err := a.verifyRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, err
	}

	newRefreshToken, err := newAuthToken(authTokenCreationFields{*a, s.Id, true})
	if err != nil {
		return nil, nil, err
	}

	s.updateRefreshToken(*newRefreshToken)

	newAccessToken, err := newAuthToken(authTokenCreationFields{*a, s.Id, false})
	if err != nil {
		return nil, nil, err
	}

	return newAccessToken, newRefreshToken, nil
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

// Return sessions that have just been updated
func (a *Account) SessionsToPersist() []Session {
	sessions := []Session{}
	for _, s := range a.ActiveSessions {
		if s.UpdatedAt.After(time.Now().Add(time.Duration(-5) * time.Second)) {
			sessions = append(sessions, s)
		}
	}

	if a.justTerminatedSessions != nil {
		sessions = append(sessions, a.justTerminatedSessions...)
	}

	return sessions
}

// Check if access token is valid and return its related session
func (a *Account) VerifyAccessToken(token *authToken) (*Session, error) {
	if token.IsRefresh() {
		return nil, normalizederr.NewRequestError("Non access token.", "")
	}

	return a.verifyToken(token)
}

// Check if refresh token is valid and return its related session
func (a *Account) verifyRefreshToken(token *authToken) (*Session, error) {
	if !token.IsRefresh() {
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

func (a *Account) verifyToken(token *authToken) (*Session, error) {
	if token.IsExpired() {
		return nil, normalizederr.NewUnauthorizedError("Expired token")
	}

	s := a.session(token.Claims.SessionId)
	if s == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid session")
	}

	if token.Claims.Sub != a.Id || token.Claims.Aud != s.Application.Id {
		return nil, normalizederr.NewUnauthorizedError("Mismatched authentication")
	}

	return s, nil
}

// Return desired active session or nil if it does not exist
func (a *Account) session(sessionId uuid.UUID) *Session {
	for _, s := range a.ActiveSessions {
		if s.Id == sessionId && s.IsActive {
			return &s
		}
	}

	return nil
}

// Return active sessions of the desired application
func (a *Account) sessionsByApp(app Application) []Session {
	sessions := []Session{}
	for _, s := range a.ActiveSessions {
		if s.Application.Id == app.Id && s.IsActive {
			sessions = append(sessions, s)
		}
	}

	return sessions
}
