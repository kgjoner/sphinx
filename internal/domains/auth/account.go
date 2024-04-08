package auth

import (
	"fmt"
	"regexp"
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
	"golang.org/x/crypto/sha3"
)

type Account struct {
	InternalId int                `json:"-"`
	Id         uuid.UUID          `json:"id" validate:"required"`
	Email      htypes.Email       `json:"-" validate:"required"`
	Phone      htypes.PhoneNumber `json:"-"`
	Password   string             `json:"-" validate:"required"`
	Username   string             `json:"username" validate:"wordId,atLeastOne=letter"`
	Document   htypes.Document    `json:"-"`

	IsActive             bool                       `json:"isActive"`
	HasEmailBeenVerified bool                       `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool                       `json:"hasPhoneBeenVerified"`
	Codes                map[AccountCodeKind]string `json:"-"`
	Links                []Link                     `json:"-"`
	ActiveSessions       []Session                  `json:"-"`

	justTerminatedSessions []Session  `json:"-"`
	AuthedSession          *Session   `json:"-"`
	AuthToken              *authToken `json:"-"`

	PasswordUpdatedAt time.Time `json:"-"`
	CreatedAt         time.Time `json:"createdAt" validate:"required"`
	UpdatedAt         time.Time `json:"updatedAt" validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type AccountCreationFields struct {
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone"`
	Password string             `json:"password" validate:"required"`
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
		Password: hashPassword(a.Password),
		Username: strings.ToLower(a.Username),
		Document: a.Document,

		IsActive:  true,
		Codes:     map[AccountCodeKind]string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	acc.generateCodeFor(AccountCodeKindValues.EMAIL_VERIFICATION)
	if !acc.Phone.IsZero() {
		acc.generateCodeFor(AccountCodeKindValues.PHONE_VERIFICATION)
	}

	return acc, validator.Validate(acc)
}

func hashPassword(str string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(str), 14)
	return string(hash)
}

func hashData(str string) string {
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, []byte(str))
	return fmt.Sprintf("%x", hash)
}

func validatePasswordInput(password string) error {
	err := validator.Validate(password, "required", "min=8", "max=32", "atLeastOne=letter number specialChar")
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("Password: %v", err.Error())
	return normalizederr.NewValidationError(msg)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (a Account) Name() string {
	if a.Username != "" {
		return a.Username
	}

	pattern := regexp.MustCompile("^(.+)@")
	matches := pattern.FindStringSubmatch(a.Email.String())
	return matches[1]
}

func (a *Account) ChangePassword(oldPassword string, newPassword string) error {
	if !a.DoesPasswordMatch(oldPassword) {
		return normalizederr.NewUnauthorizedError("Invalid credentials.")
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
		return normalizederr.NewUnauthorizedError("Invalid code.")
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

func (a Account) DoesPasswordMatch(password string) bool {
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

// Check whether account has role on application related to their authed session
func (a Account) HasRoleOnAuth(roles ...Role) bool {
	link := a.authedLink()
	if link == nil {
		return false
	}

	return link.hasRole(roles...)
}

// Check whether account has role on a target application
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
	app := actor.AuthedSession.Application
	if !actor.HasRole(app, RoleValues.ADMIN) {
		return normalizederr.NewUnauthorizedError("Does not have permission to execute this action.")
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

	return a.link(a.AuthedSession.Application)
}

/* ==============================================================================
	SESSION RELATED METHODS
============================================================================== */

// Create a new session and generate tokens
func (a *Account) InitSession(password string, f *SessionCreationFields) (access *authToken, refresh *authToken, err error) {
	if !a.DoesPasswordMatch(password) {
		return nil, nil, normalizederr.NewUnauthorizedError("Invalid credentials.")
	}

	// Check if link exists
	if link := a.link(f.Application); link == nil {
		err := a.LinkTo(f.Application)
		if err != nil {
			return nil, nil, err
		}
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
	a.ActiveSessions = append(a.ActiveSessions, *session)

	refreshToken, err := newAuthToken(authTokenCreationFields{*a, session.Id, true})
	if err != nil {
		return nil, nil, err
	}

	session.updateRefreshToken(*refreshToken)
	a.ActiveSessions[len(a.ActiveSessions)-1] = *session

	accessToken, err := newAuthToken(authTokenCreationFields{*a, session.Id, false})
	if err != nil {
		return nil, nil, err
	}

	a.AuthedSession = session
	a.AuthToken = accessToken
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
			a.TerminateAllSessions()
			return normalizederr.NewFatalUnauthorizedError("Revoked token.")
		}
	}

	a.AuthToken = token
	a.AuthedSession = s
	return nil
}

func (a Account) IsAuthenticated() bool {
	return a.AuthToken != nil
}

// Generate new tokens for an existing session
func (a *Account) IssueNewTokens() (access *authToken, refresh *authToken, err error) {
	if !a.IsAuthenticated() {
		return nil, nil, normalizederr.NewUnauthorizedError("Unauthenticated.")
	} else if !a.AuthToken.IsRefresh() {
		return nil, nil, normalizederr.NewUnauthorizedError("Non refresh token")
	}

	newRefreshToken, err := newAuthToken(authTokenCreationFields{*a, a.AuthedSession.Id, true})
	if err != nil {
		return nil, nil, err
	}

	a.AuthedSession.updateRefreshToken(*newRefreshToken)

	newAccessToken, err := newAuthToken(authTokenCreationFields{*a, a.AuthedSession.Id, false})
	if err != nil {
		return nil, nil, err
	}

	a.AuthToken = newAccessToken

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

			if a.AuthedSession != nil && sessionId == a.AuthedSession.Id {
				a.AuthedSession = nil
				a.AuthToken = nil
			}

			return &s, nil
		}
	}

	return nil, normalizederr.NewRequestError("Session not found.", "")
}

func (a *Account) TerminateAuthedSession() (*Session, error) {
	if !a.IsAuthenticated() {
		return nil, normalizederr.NewUnauthorizedError("Unauthenticated")
	}

	s, err := a.TerminateSession(a.AuthedSession.Id)
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
	a.AuthToken = nil
	a.AuthedSession = nil
	return nil
}

// Return sessions that have just been updated
func (a *Account) SessionsToPersist() []Session {
	sessions := []Session{}
	for _, s := range a.ActiveSessions {
		if a.AuthedSession != nil && a.AuthedSession.Id == s.Id {
			continue
		}

		if s.UpdatedAt.After(time.Now().Add(time.Duration(-5) * time.Second)) {
			sessions = append(sessions, s)
		}
	}

	if a.justTerminatedSessions != nil {
		sessions = append(sessions, a.justTerminatedSessions...)
	}

	if a.AuthedSession != nil && a.AuthedSession.UpdatedAt.After(time.Now().Add(time.Duration(-5)*time.Second)) {
		sessions = append(sessions, *a.AuthedSession)
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
	Id       uuid.UUID          `json:"id" validate:"required"`
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty" validate:"wordId"`
	Document htypes.Document    `json:"document,omitempty"`

	IsActive             bool  `json:"isActive"`
	HasEmailBeenVerified bool  `json:"hasEmailBeenVerified"`
	HasPhoneBeenVerified bool  `json:"hasPhoneBeenVerified"`
	Link                 *Link `json:"link"`
}

func (a Account) PrivateView(actor Account) (*AccountPrivateView, error) {
	if a.Id != actor.Id && actor.HasRoleOnAuth(RoleValues.ADMIN) {
		return nil, normalizederr.NewForbiddenError("Does not have permission to view this user information.")
	}

	link := actor.authedLink()
	if a.Id != actor.Id {
		link = a.link(link.Application)
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
		link,
	}, nil
}
