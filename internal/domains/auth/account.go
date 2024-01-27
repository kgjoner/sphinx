package auth

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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
	Username   string             `json:"username" validator:"wordId,atLeastOne=letter"`
	Document   htypes.Document    `json:"document"`

	IsActive             bool                       `json:"isActive"`
	HasEmailBeenVerified bool                       `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool                       `json:"hasPhoneBeenVerified"`
	Codes                map[AccountCodeKind]string `json:"-"`
	Links                []Link                     `json:"-"`
	ActiveSessions       []Session                  `json:"-"`

	justTerminatedSessions []Session  `json:"-"`
	authedSession          *Session   `json:"-"`
	authToken              *authToken `json:"-"`

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
	err := validatePasswordInput(a.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	acc := &Account{
		Id:       uuid.New(),
		Email:    a.Email.Normalize(),
		Phone:    a.Phone,
		Password: hashData(a.Password),
		Username: strings.ToLower(a.Username),
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

func validatePasswordInput(password string) error {
	err := validator.Validate(password, "min=8,atLeastOne=letter,number,specialChar")
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("Password: %v", err.Error())
	return errors.New(msg)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (a *Account) ChangePassword(oldPassword string, newPassword string) error {
	if !a.doesPasswordMatch(oldPassword) {
		return normalizederr.NewForbiddenError("Invalid credentials.")
	}

	err := validatePasswordInput(newPassword)
	if err != nil {
		return err
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

	err := validatePasswordInput(newPassword)
	if err != nil {
		return err
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

	if f.Username != "" {
		f.Username = strings.ToLower(f.Username)
	}

	structop.New(a).Update(f)
	a.UpdatedAt = time.Now()

	if !f.Phone.IsZero() {
		a.generateCodeFor(AccountCodeKindValues.PHONE_VERIFICATION)
	}
	return validator.Validate(a)
}

func (a Account) doesPasswordMatch(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
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

func (a *Account) AddRole(r Role, actor Account) error {
	return a.updatePermission(actor, func(l *Link) error {
		return l.addRole(r)
	})
}

func (a *Account) RemoveRole(r Role, actor Account) error {
	return a.updatePermission(actor, func(l *Link) error {
		return l.removeRole(r)
	})
}

func (a *Account) AddGranting(g string, actor Account) error {
	return a.updatePermission(actor, func(l *Link) error {
		return l.addGranting(g)
	})
}

func (a *Account) RemoveGranting(g string, actor Account) error {
	return a.updatePermission(actor, func(l *Link) error {
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
func (a *Account) updatePermission(actor Account, updaterFn func(*Link) error) error {
	app := actor.authedSession.Application
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

// Return link related to auth token or nil if account is not authenticated
func (a *Account) authedLink() *Link {
	if !a.IsAuthenticated() {
		return nil
	}

	return a.link(a.authedSession.Application)
}

/* ==============================================================================
	SESSION RELATED METHODS
============================================================================== */

// Create a new session and generate tokens
func (a *Account) InitSession(password string, f *SessionCreationFields) (access *authToken, refresh *authToken, err error) {
	if !a.doesPasswordMatch(password) {
		return nil, nil, normalizederr.NewForbiddenError("Invalid credentials.")
	}

	// Check if link exists
	if link := a.link(f.Application); link == nil {
		err := a.LinkTo(f.Application)
		if err != nil {
			return nil, nil, err
		}
	}

	// Terminate exceeding sessions
	if conauthedSessions :=
		SessionSortableByAge(a.sessionsByApp(f.Application)); config.Environment.MAX_CONCURRENT_SESSIONS > 0 &&
		len(conauthedSessions) >= config.Environment.MAX_CONCURRENT_SESSIONS {

		previousExcess := len(conauthedSessions) - config.Environment.MAX_CONCURRENT_SESSIONS
		/* If all goes well, previousExcess will be zero. There will be only one session to terminate
		for preventing new session to exceed max limit */
		sort.Sort(conauthedSessions)
		sessionsToTerminate := conauthedSessions[:(previousExcess + 1)]
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
	a.authedSession = session
	a.authToken = accessToken
	return accessToken, refreshToken, validator.Validate(a)
}

func (a *Account) Authenticate(token *authToken) error {
	s, err := a.verifyToken(token)
	if err != nil {
		return err
	}

	if token.IsRefresh() {
		doesMatch := s.doesRefreshTokenMatch(token.String())
		if !doesMatch {
			return normalizederr.NewUnauthorizedError("Revoked token.")
		}
	}

	a.authToken = token
	a.authedSession = s
	return nil
}

func (a Account) IsAuthenticated() bool {
	return a.authToken != nil
}

// Generate new tokens for an existing session
func (a *Account) IssueNewTokens() (access *authToken, refresh *authToken, err error) {
	if !a.IsAuthenticated() {
		return nil, nil, normalizederr.NewUnauthorizedError("Unauthenticated.")
	} else if !a.authToken.IsRefresh() {
		return nil, nil, normalizederr.NewForbiddenError("Non refresh token")
	}

	newRefreshToken, err := newAuthToken(authTokenCreationFields{*a, a.authedSession.Id, true})
	if err != nil {
		return nil, nil, err
	}

	a.authedSession.updateRefreshToken(*newRefreshToken)

	newAccessToken, err := newAuthToken(authTokenCreationFields{*a, a.authedSession.Id, false})
	if err != nil {
		return nil, nil, err
	}

	a.authToken = newAccessToken

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

			if a.authedSession != nil && sessionId == a.authedSession.Id {
				a.authedSession = nil
				a.authToken = nil
			}

			return &s, nil
		}
	}

	return nil, normalizederr.NewRequestError("Session not found.", "")
}

func (a *Account) TerminateAuthedSession() (*Session, error) {
	if a.IsAuthenticated() {
		return nil, normalizederr.NewUnauthorizedError("Unauthenticated")
	}

	s, err := a.TerminateSession(a.authedSession.Id)
	if err != nil {
		return nil, err
	}

	return s, nil
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
	a.authToken = nil
	a.authedSession = nil
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

	if a.authedSession != nil && a.authedSession.UpdatedAt.After(time.Now().Add(time.Duration(-5)*time.Second)) {
		sessions = append(sessions, *a.authedSession)
	}

	return sessions
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

/* ==============================================================================
	VIEWS
============================================================================== */

type AccountPrivateView struct {
	Id       uuid.UUID          `json:"id" validator:"required"`
	Email    htypes.Email       `json:"email" validator:"required"`
	Phone    htypes.PhoneNumber `json:"phone"`
	Username string             `json:"username" validator:"wordId"`
	Document htypes.Document    `json:"document"`

	IsActive             bool  `json:"isActive"`
	HasEmailBeenVerified bool  `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool  `json:"hasPhoneBeenVerified"`
	Link                 *Link `json:"link"`
}

func (a Account) PrivateView(actor Account) *AccountPrivateView {
	app := actor.authedSession.Application
	if a.Id != actor.Id && actor.HasRole(app, RoleValues.ADMIN) {
		return nil
	}

	return &AccountPrivateView{
		a.Id,
		a.Email,
		a.Phone,
		a.Username,
		a.Document,
		a.IsActive,
		a.HasEmailBeenVerified,
		a.HasPhoneBeenVerified,
		a.authedLink(),
	}
}
