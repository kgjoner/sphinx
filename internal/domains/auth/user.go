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
	"github.com/kgjoner/cornucopia/utils/pwdgen"
	"github.com/kgjoner/cornucopia/utils/sliceman"
	"github.com/kgjoner/cornucopia/utils/structop"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

type User struct {
	InternalID int                `json:"-"`
	ID         uuid.UUID          `json:"id" validate:"required"`
	Email      htypes.Email       `json:"-" validate:"required"`
	Phone      htypes.PhoneNumber `json:"-"`
	Password   string             `json:"-" validate:"required"`
	Username   string             `json:"username" validate:"wordID,atLeastOne=letter"`
	Document   htypes.Document    `json:"-"`
	ExtraData  `json:"extraData,omitempty"`

	IsActive             bool                        `json:"isActive"`
	PendingEmail         htypes.Email                `json:"-"`
	HasEmailBeenVerified bool                        `json:"-"`
	PendingPhone         htypes.PhoneNumber          `json:"-"`
	HasPhoneBeenVerified bool                        `json:"-"`
	VerificationCodes    map[VerificationKind]string `json:"-"`
	Links                []Link                      `json:"-"`
	ActiveSessions       []Session                   `json:"-"`

	justTerminatedSessions []Session  `json:"-"`
	hasValidCredentials    bool       `json:"-"`
	AuthedSession          *Session   `json:"-"`
	AuthToken              *authToken `json:"-"`

	PasswordUpdatedAt htypes.NullTime `json:"-"`
	UsernameUpdatedAt htypes.NullTime `json:"-"`
	CreatedAt         time.Time       `json:"createdAt" validate:"required"`
	UpdatedAt         time.Time       `json:"updatedAt" validate:"required"`
}

type ExtraData struct {
	Name    string         `json:"name,omitempty"`
	Surname string         `json:"surname,omitempty"`
	Address htypes.Address `json:"-"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type UserCreationFields struct {
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone"`
	Password string             `json:"password" validate:"required"`
	Username string             `json:"username"`
	Document htypes.Document    `json:"document"`
	UserExtraFields
}

func NewUser(a *UserCreationFields) (*User, error) {
	err := validatePasswordInput(a.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	acc := &User{
		ID:        uuid.New(),
		Email:     a.Email,
		Phone:     a.Phone,
		Password:  hashPassword(a.Password),
		Username:  strings.ToLower(a.Username),
		Document:  a.Document,
		ExtraData: ExtraData(a.UserExtraFields),

		IsActive:          true,
		VerificationCodes: map[VerificationKind]string{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// TODO: Add link for root Application as default

	acc.generateCodeFor(VerificationEmail)
	if !acc.Phone.IsZero() {
		acc.generateCodeFor(VerificationPhone)
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
	// err := validator.Validate(password, "required", "min=8", "max=32", "atLeastOne=letter number specialChar")
	err := validator.Validate(password, "required", "min=8", "max=32", "atLeastOne=letter number")
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("Password: %v", err.Error())
	return normalizederr.NewValidationError(msg)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (a User) Name() string {
	if a.ExtraData.Name != "" {
		return a.ExtraData.Name
	}

	if a.Username != "" {
		return a.Username
	}

	pattern := regexp.MustCompile("^(.+)@")
	matches := pattern.FindStringSubmatch(a.Email.String())
	return matches[1]
}

func (a *User) VerifyUser(kind VerificationKind, code string) error {
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
		if a.HasEmailBeenVerified && a.PendingEmail.IsZero() {
			return normalizederr.NewRequestError("Email has already been verified.")
		}
	case VerificationPhone:
		if a.HasPhoneBeenVerified && a.PendingPhone.IsZero() {
			return normalizederr.NewRequestError("Phone has already been verified.")
		}
	}

	if a.VerificationCodes[kind] == "" {
		return normalizederr.NewConflictError("User does not have a verification code.")
	}

	if code != a.VerificationCodes[kind] {
		return normalizederr.NewRequestError("Invalid code.")
	}

	switch kind {
	case VerificationEmail:
		if !a.PendingEmail.IsZero() {
			// Confirm email update
			a.Email = a.PendingEmail
			var emptyEmail htypes.Email
			a.PendingEmail = emptyEmail
		}
		a.HasEmailBeenVerified = true
	case VerificationPhone:
		if !a.PendingPhone.IsZero() {
			// Confirm phone update
			a.Phone = a.PendingPhone
			var emptyPhone htypes.PhoneNumber
			a.PendingPhone = emptyPhone
		}
		a.HasPhoneBeenVerified = true
	}
	a.clearCodeFor(kind)
	a.UpdatedAt = time.Now()

	return nil
}

func (a *User) ChangePassword(oldPassword string, newPassword string) error {
	if !a.DoesPasswordMatch(oldPassword) {
		return normalizederr.NewUnauthorizedError("Invalid credentials.")
	}

	err := validatePasswordInput(newPassword)
	if err != nil {
		return err
	}

	a.Password = hashPassword(newPassword)
	a.TerminateAllSessions()

	now := time.Now()
	a.PasswordUpdatedAt = htypes.NullTime{Time: now}
	a.UpdatedAt = now
	return validator.Validate(a)
}

func (a *User) RequestPasswordReset() (string, error) {
	a.generateCodeFor(VerificationPasswordReset)
	a.UpdatedAt = time.Now()
	return a.VerificationCodes[VerificationPasswordReset], validator.Validate(a)
}

func (a *User) ResetPassword(newPassword string, code string) error {
	resetCode := a.VerificationCodes[VerificationPasswordReset]
	if resetCode == "" {
		return normalizederr.NewRequestError("User did not request a password reset.", "")
	} else if code != a.VerificationCodes[VerificationPasswordReset] {
		return normalizederr.NewRequestError("Invalid code.", "")
	}

	err := validatePasswordInput(newPassword)
	if err != nil {
		return err
	}

	a.Password = hashPassword(newPassword)
	a.clearCodeFor(VerificationPasswordReset)
	a.TerminateAllSessions()

	now := time.Now()
	a.PasswordUpdatedAt = htypes.NullTime{Time: now}
	a.UpdatedAt = now
	return validator.Validate(a)
}

type UserMissableFields struct {
	Phone    htypes.PhoneNumber `json:"phone"`
	Username string             `json:"username"`
	Document htypes.Document    `json:"document"`
}

func (a *User) AddMissingFields(f UserMissableFields) error {
	if !f.Phone.IsZero() && !a.Phone.IsZero() {
		return normalizederr.NewRequestError("Phone was already added.", "")
	}
	if f.Username != "" && a.Username != "" {
		return normalizederr.NewRequestError("Username was already added.", "")
	}
	if !f.Document.IsZero() && !a.Document.IsZero() {
		return normalizederr.NewRequestError("Document was already added.", "")
	}

	now := time.Now()
	if f.Username != "" {
		f.Username = strings.ToLower(f.Username)
		a.UsernameUpdatedAt = htypes.NullTime{Time: now}
	}

	structop.New(a).Update(f)
	a.UpdatedAt = now

	if !f.Phone.IsZero() {
		a.generateCodeFor(VerificationPhone)
	}
	return validator.Validate(a)
}

type UserUniqueFields struct {
	Email    htypes.Email       `json:"email"`
	Phone    htypes.PhoneNumber `json:"phone"`
	Username string             `json:"username"`
	Document htypes.Document    `json:"document"`
}

func (a *User) UpdateUniqueFields(f UserUniqueFields) error {
	if !f.Email.IsZero() {
		a.PendingEmail = f.Email
		a.generateCodeFor(VerificationEmail)
	}

	if !f.Phone.IsZero() {
		a.PendingPhone = f.Phone
		a.generateCodeFor(VerificationPhone)
	}

	now := time.Now()
	if f.Username != "" {
		if a.UsernameUpdatedAt.Time.After(now.Add(-time.Hour * 24 * 90)) {
			return normalizederr.NewRequestError("Username can only be updated once every 90 days.")
		}

		a.Username = strings.ToLower(f.Username)
		a.UsernameUpdatedAt = htypes.NullTime{Time: now}
	}

	if !f.Document.IsZero() {
		a.Document = f.Document
	}

	a.UpdatedAt = now
	return validator.Validate(a)
}

type UserExtraFields struct {
	Name    string         `json:"name"`
	Surname string         `json:"surname"`
	Address htypes.Address `json:"address"`
}

func (a *User) UpdateExtraData(f UserExtraFields) error {
	structop.New(&a.ExtraData).Update(f)
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

// Allows users to cancel their pending field update (email or phone).
// It will remove the pending field and clear the code for it.
func (a *User) CancelPendingField(field string) error {
	err := validator.Validate(field, "required", "oneof=email phone")
	if err != nil {
		return err
	}

	switch field {
	case "email":
		return a.cancelPendingEmail()
	case "phone":
		return a.cancelPendingPhone()
	default:
		return normalizederr.NewRequestError("Invalid field name.")
	}
}

// Allows users to cancel their pending email update
func (a *User) cancelPendingEmail() error {
	if a.PendingEmail.IsZero() {
		return normalizederr.NewRequestError("No pending email update to cancel.")
	}

	var emptyEmail htypes.Email
	a.PendingEmail = emptyEmail
	a.clearCodeFor(VerificationEmail)
	a.UpdatedAt = time.Now()

	return nil
}

// Allows users to cancel their pending phone update
func (a *User) cancelPendingPhone() error {
	if a.PendingPhone.IsZero() {
		return normalizederr.NewRequestError("No pending phone update to cancel.")
	}

	var emptyPhone htypes.PhoneNumber
	a.PendingPhone = emptyPhone
	a.clearCodeFor(VerificationPhone)
	a.UpdatedAt = time.Now()

	return nil
}

func (a User) DoesPasswordMatch(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// It will generate an random string and save it in the field Code, under the desired key (kind).
// It overwrites old value, if any.
func (a *User) generateCodeFor(kind VerificationKind) {
	a.VerificationCodes[kind] = pwdgen.Generate(12, "lower", "upper", "number")
}

func (a *User) clearCodeFor(kind VerificationKind) {
	delete(a.VerificationCodes, kind)
}

/* ==============================================================================
	LINK RELATED METHODS
============================================================================== */

// Agrees with link to desired application. If there is no link, it will create a new one.
func (a *User) GiveConsent(app Application) error {
	for i, l := range a.Links {
		if l.Application.ID == app.ID {
			if l.HasConsent {
				return normalizederr.NewRequestError("User has already given link to desired application.", "")
			}

			return a.Links[i].restoreConsent()
		}
	}

	link := newLink(a, app)
	a.Links = append(a.Links, *link)
	return validator.Validate(a)
}

// Revoke consent in linking with desired application.
func (a *User) RevokeConsent(app Application) error {
	for i, l := range a.Links {
		if l.Application.ID == app.ID {
			if !l.HasConsent {
				return normalizederr.NewRequestError("User has already revoked link to desired application.", "")
			}

			return a.Links[i].revokeConsent()
		}
	}

	return normalizederr.NewRequestError("User has not given link to desired application.", "")
}

// Check whether user has role on application related to their authed session.
// False is also returned when user is not authed.
func (a User) HasRoleOnAuth(roles ...Role) bool {
	link := a.authedLink()
	if link == nil || !link.HasConsent {
		return false
	}

	return link.hasRole(roles...)
}

// Check whether user has role on a target application.
func (a User) HasRole(app Application, roles ...Role) bool {
	link := a.link(app.ID)
	if link == nil || !link.HasConsent {
		return false
	}

	return link.hasRole(roles...)
}

func (a *User) AddRole(r Role, app Application) error {
	return a.updatePermission(app, func(c *Link) error {
		return c.addRole(r)
	})
}

func (a *User) RemoveRole(r Role, app Application) error {
	return a.updatePermission(app, func(c *Link) error {
		return c.removeRole(r)
	})
}

// Return application links that have just been updated
func (a *User) LinksToPersist() []Link {
	links := []Link{}
	for _, l := range a.Links {
		if l.UpdatedAt.After(time.Now().Add(time.Duration(-5) * time.Second)) {
			links = append(links, l)
		}
	}

	return links
}

// Prepare and execute desired role updates
func (a *User) updatePermission(app Application, updaterFn func(*Link) error) error {
	if !app.IsAuthenticated() {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	link := a.link(app.ID)
	if link == nil || !link.HasConsent {
		return normalizederr.NewForbiddenError("target user has not consented to desired application")
	}

	err := updaterFn(link)
	if err != nil {
		return err
	}

	for i, c := range a.Links {
		if c.ID == link.ID {
			a.Links[i] = *link
		}
	}

	return nil
}

// Return desired application link or nil if user has never given link to it.
func (a *User) link(appID uuid.UUID) *Link {
	for i, c := range a.Links {
		if c.Application.ID == appID {
			return &a.Links[i]
		}
	}

	return nil
}

// Return application link related to auth token or nil if user is not authenticated
func (a *User) authedLink() *Link {
	if !a.IsAuthenticated() {
		return nil
	}

	return a.link(a.AuthedSession.Application.ID)
}

/* ==============================================================================
	SESSION RELATED METHODS
============================================================================== */

func (a *User) Authenticate(token *authToken) error {
	s, err := a.verifyToken(token)
	if err != nil {
		return err
	}

	if token.IsRefresh() {
		doesMatch := s.doesRefreshTokenMatch(token.String())
		if !doesMatch {
			a.TerminateAllSessions()
			return normalizederr.NewFatalUnauthorizedError("revoked token", errcode.ExpiredSession)
		}
	}

	link := a.link(s.Application.ID)
	if link == nil || !link.HasConsent {
		return normalizederr.NewForbiddenError("target user has revoked consent to desired application", errcode.RevokedConsent)
	}

	if !a.IsActive {
		return normalizederr.NewForbiddenError("user is no longer active", errcode.DeactivatedUser)
	}

	a.AuthToken = token
	a.AuthedSession = s
	return nil
}

func (a *User) AuthenticateViaPassword(password string) error {
	if !a.DoesPasswordMatch(password) {
		return normalizederr.NewUnauthorizedError("Invalid credentials.", errcode.InvalidCredentials)
	}

	if !a.IsActive {
		return normalizederr.NewForbiddenError("user is no longer active", errcode.DeactivatedUser)
	}

	a.hasValidCredentials = true
	return nil
}

func (a *User) AuthenticateViaGrant(grant *AuthorizationGrant, credentials *GrantCredentials) error {
	if err := grant.use(*a, credentials); err != nil {
		return err
	}

	if !a.IsActive {
		return normalizederr.NewForbiddenError("user is no longer active", errcode.DeactivatedUser)
	}

	a.hasValidCredentials = true
	return nil
}

func (a User) IsAuthenticated() bool {
	return a.AuthToken != nil || a.hasValidCredentials
}

// Creates an authorization grant for an application that user has linked to
func (a User) IssueAuthorizationGrant(f *AuthorizationGrantCreationFields, appID uuid.UUID) (*AuthorizationGrant, error) {
	if !a.IsAuthenticated() {
		return nil, normalizederr.NewUnauthorizedError("User must be authenticated.")
	}

	if a.AuthedSession != nil && !a.AuthedSession.Application.isRoot() {
		return nil, normalizederr.NewForbiddenError("User can only issue grants with a root app auth token")
	}

	return newAuthorizationGrant(a, f)
}

// Create a new session and generate tokens
func (a *User) InitSession(f *SessionCreationFields) (access *authToken, refresh *authToken, err error) {
	if !a.IsAuthenticated() {
		return nil, nil, normalizederr.NewUnauthorizedError("User must be authenticated.")
	}

	if !a.hasValidCredentials {
		return nil, nil, normalizederr.NewForbiddenError("User can only init sessions with valid credentials, not with another auth token.")
	}

	// Check if link exists and is active
	link := a.link(f.Application.ID)
	if link == nil || !link.HasConsent {
		return nil, nil, normalizederr.NewForbiddenError("User has not consented to this application.")
	}
	f.Application = link.Application

	// Terminate exceeding sessions
	if concurrentSessions :=
		SessionSortableByAge(a.sessionsByApp(f.Application)); config.Env.MAX_CONCURRENT_SESSIONS > 0 &&
		len(concurrentSessions) >= config.Env.MAX_CONCURRENT_SESSIONS {

		previousExcess := len(concurrentSessions) - config.Env.MAX_CONCURRENT_SESSIONS
		/* If all goes well, previousExcess will be zero. There will be only one session to terminate
		for preventing new session to exceed max limit */
		sort.Sort(concurrentSessions)
		sessionsToTerminate := concurrentSessions[:(previousExcess + 1)]
		for _, s := range sessionsToTerminate {
			_, err := a.TerminateSession(s.ID)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	// Create session and tokens
	session := newSession(*a, f)
	a.ActiveSessions = append(a.ActiveSessions, *session)

	refreshToken, err := newAuthToken(authTokenCreationFields{*a, session.ID, true})
	if err != nil {
		return nil, nil, err
	}

	session.updateRefreshToken(*refreshToken)
	a.ActiveSessions[len(a.ActiveSessions)-1] = *session

	accessToken, err := newAuthToken(authTokenCreationFields{*a, session.ID, false})
	if err != nil {
		return nil, nil, err
	}

	a.AuthedSession = session
	a.AuthToken = accessToken
	return accessToken, refreshToken, validator.Validate(a)
}

// Generate new tokens for an existing session
func (a *User) IssueNewTokens() (access *authToken, refresh *authToken, err error) {
	if !a.IsAuthenticated() {
		return nil, nil, normalizederr.NewUnauthorizedError("Unauthenticated.")
	} else if !a.AuthToken.IsRefresh() {
		return nil, nil, normalizederr.NewUnauthorizedError("Non refresh token", errcode.InvalidAccess)
	}

	newRefreshToken, err := newAuthToken(authTokenCreationFields{*a, a.AuthedSession.ID, true})
	if err != nil {
		return nil, nil, err
	}

	a.AuthedSession = a.session(a.AuthedSession.ID)
	if a.AuthedSession == nil {
		return nil, nil, normalizederr.NewConflictError("No active session.", errcode.SessionNotFound)
	}

	a.AuthedSession.updateRefreshToken(*newRefreshToken)

	newAccessToken, err := newAuthToken(authTokenCreationFields{*a, a.AuthedSession.ID, false})
	if err != nil {
		return nil, nil, err
	}

	a.AuthToken = newAccessToken

	return newAccessToken, newRefreshToken, nil
}

func (a *User) TerminateSession(sessionID uuid.UUID) (*Session, error) {
	for i, s := range a.ActiveSessions {
		if s.ID == sessionID {
			err := s.terminate()
			if err != nil {
				return nil, err
			}

			if a.justTerminatedSessions == nil {
				a.justTerminatedSessions = []Session{}
			}
			a.justTerminatedSessions = append(a.justTerminatedSessions, s)
			a.ActiveSessions = sliceman.Remove(a.ActiveSessions, i)

			if a.AuthedSession != nil && sessionID == a.AuthedSession.ID {
				a.AuthedSession = nil
				a.AuthToken = nil
			}

			return &s, nil
		}
	}

	return nil, normalizederr.NewRequestError("Session not found.", errcode.SessionNotFound)
}

func (a *User) TerminateAuthedSession() (*Session, error) {
	if !a.IsAuthenticated() {
		return nil, normalizederr.NewUnauthorizedError("Unauthenticated")
	}

	s, err := a.TerminateSession(a.AuthedSession.ID)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (a *User) TerminateAllSessions() error {
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
func (a *User) SessionsToPersist() []Session {
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

func (a *User) verifyToken(token *authToken) (*Session, error) {
	if token.IsExpired() {
		code := errcode.ExpiredAccess
		if token.IsRefresh() {
			code = errcode.ExpiredSession
		}
		return nil, normalizederr.NewUnauthorizedError("Expired token", code)
	}

	s := a.session(token.Claims.SessionID)
	if s == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid session", errcode.InvalidAccess)
	}

	if token.Claims.Sub != a.ID || token.Claims.Aud != s.Application.ID {
		return nil, normalizederr.NewUnauthorizedError("Mismatched authentication", errcode.InvalidAccess)
	}

	return s, nil
}

// Return desired active session or nil if it does not exist
func (a *User) session(sessionID uuid.UUID) *Session {
	for i, s := range a.ActiveSessions {
		if s.ID == sessionID && s.IsActive {
			return &a.ActiveSessions[i]
		}
	}

	return nil
}

// Return active sessions of the desired application
func (a *User) sessionsByApp(app Application) []Session {
	sessions := []Session{}
	for _, s := range a.ActiveSessions {
		if s.Application.ID == app.ID && s.IsActive {
			sessions = append(sessions, s)
		}
	}

	return sessions
}

/* ==============================================================================
	VIEWS
============================================================================== */

type UserPrivateView struct {
	ID       uuid.UUID          `json:"id" validate:"required"`
	Email    htypes.Email       `json:"email" validate:"required"`
	Phone    htypes.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty"`
	Document htypes.Document    `json:"document,omitempty"`
	Name     string             `json:"name,omitempty"`
	Surname  string             `json:"surname,omitempty"`
	Address  *htypes.Address    `json:"address,omitempty"`

	IsActive             bool               `json:"isActive"`
	PendingEmail         htypes.Email       `json:"pendingEmail,omitempty"`
	HasEmailBeenVerified bool               `json:"hasEmailBeenVerified"`
	PendingPhone         htypes.PhoneNumber `json:"pendingPhone,omitempty"`
	HasPhoneBeenVerified bool               `json:"hasPhoneBeenVerified"`
	Link                 *Link              `json:"link,omitempty"`
}

func (a User) PrivateView(actor User) (*UserPrivateView, error) {
	if !(a.ID == actor.ID || actor.HasRoleOnAuth(RoleAdmin)) {
		return nil, normalizederr.NewForbiddenError("Does not have permission to view this user information.")
	}

	link := actor.authedLink()
	if a.ID != actor.ID {
		link = a.link(link.Application.ID)
	}

	var address *htypes.Address
	if a.ExtraData.Address != (htypes.Address{}) {
		address = &a.ExtraData.Address
	}

	return &UserPrivateView{
		a.ID,
		a.Email,
		a.Phone,
		a.Username,
		a.Document,
		a.ExtraData.Name,
		a.ExtraData.Surname,
		address,
		a.IsActive,
		a.PendingEmail,
		a.HasEmailBeenVerified,
		a.PendingPhone,
		a.HasPhoneBeenVerified,
		link,
	}, nil
}
