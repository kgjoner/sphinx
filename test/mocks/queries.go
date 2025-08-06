package mocks

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

// MockQueries provides a mock implementation of database queries
type MockQueries struct {
	mu           sync.RWMutex
	users        map[uuid.UUID]*auth.User
	sessions     map[uuid.UUID]*auth.Session
	applications map[uuid.UUID]*auth.Application
	links        map[uuid.UUID]*auth.Link

	// Mock database errors for testing error scenarios
	shouldError   bool
	errorToReturn error
}

func NewMockQueries() *MockQueries {
	q := &MockQueries{
		users:        make(map[uuid.UUID]*auth.User),
		sessions:     make(map[uuid.UUID]*auth.Session),
		applications: make(map[uuid.UUID]*auth.Application),
		links:        make(map[uuid.UUID]*auth.Link),
	}

	q.InsertApplication(RootApplication)
	AdminRootLink.Application = *RootApplication
	SimpleUserRootLink.Application = *RootApplication

	q.InsertApplication(CommonApplication)

	q.InsertUser(AdminUser)
	AdminRootLink.UserID = AdminUser.InternalID
	q.InsertUser(SimpleUserUser)
	SimpleUserRootLink.UserID = SimpleUserUser.InternalID

	q.UpsertLinks(*AdminRootLink, *SimpleUserRootLink)
	return q
}

// User mock methods
func (m *MockQueries) InsertUser(acc *auth.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	// Check for unique field violations
	if err := m.checkUserUniqueConstraints(acc, nil); err != nil {
		return err
	}

	if acc.ID == uuid.Nil {
		acc.ID = uuid.New()
	}
	acc.InternalID = len(m.users) + 1
	m.users[acc.ID] = acc
	return nil
}

func (m *MockQueries) UpdateUser(acc auth.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	if _, exists := m.users[acc.ID]; !exists {
		return sql.ErrNoRows
	}

	// Check for unique field violations (exclude the user being updated)
	if err := m.checkUserUniqueConstraints(&acc, &acc.ID); err != nil {
		return err
	}

	m.users[acc.ID] = &acc
	return nil
}

// DuplicateKeyError represents a unique constraint violation
type DuplicateKeyError struct {
	Field string
	Value string
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate key value violates unique constraint on field user_%s_key: %s", e.Field, e.Value)
}

// checkUserUniqueConstraints validates that the user doesn't violate unique field constraints
// excludeID allows skipping a specific user ID (useful for updates)
func (m *MockQueries) checkUserUniqueConstraints(acc *auth.User, excludeID *uuid.UUID) error {
	for _, existing := range m.users {
		// Skip the user being updated
		if excludeID != nil && existing.ID == *excludeID {
			continue
		}

		// Check email uniqueness
		if acc.Email.String() != "" && existing.Email.String() == acc.Email.String() {
			return &DuplicateKeyError{Field: "email", Value: acc.Email.String()}
		}
		// Check username uniqueness
		if acc.Username != "" && existing.Username == acc.Username {
			return &DuplicateKeyError{Field: "username", Value: acc.Username}
		}
		// Check phone uniqueness
		if acc.Phone.String() != "" && existing.Phone.String() == acc.Phone.String() {
			return &DuplicateKeyError{Field: "phone", Value: acc.Phone.String()}
		}
		// Check document uniqueness
		if acc.Document.String() != "" && existing.Document.String() == acc.Document.String() {
			return &DuplicateKeyError{Field: "document", Value: acc.Document.String()}
		}
	}
	return nil
}

func (m *MockQueries) GetUserByID(id uuid.UUID) (*auth.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	if acc, exists := m.users[id]; exists {
		// Create a copy of the user to avoid modifying the original
		userCopy := *acc

		// Populate Links for this user
		userCopy.Links = m.getLinksForUser(acc.InternalID)

		// Populate ActiveSessions for this user (only active ones)
		userCopy.ActiveSessions = m.getActiveSessionsForUser(acc.InternalID)

		return &userCopy, nil
	}
	return nil, nil
}

func (m *MockQueries) GetUserByEntry(entry auth.Entry) (*auth.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	for _, acc := range m.users {
		if acc.Email.String() == entry.String() || acc.Username == entry.String() || acc.Document.String() == entry.String() || acc.Phone.String() == entry.String() {
			// Create a copy of the user to avoid modifying the original
			userCopy := *acc

			// Populate Links for this user
			userCopy.Links = m.getLinksForUser(acc.InternalID)

			// Populate ActiveSessions for this user (only active ones)
			userCopy.ActiveSessions = m.getActiveSessionsForUser(acc.InternalID)

			return &userCopy, nil
		}
	}
	return nil, nil
}

func (m *MockQueries) GetUserByExternalAuthID(providerName string, subjectID string) (*auth.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	for _, acc := range m.users {
		if acc.ExternalAuthIDs == nil {
			continue
		}

		if acc.ExternalAuthIDs[providerName] == subjectID {
			// Create a copy of the user to avoid modifying the original
			userCopy := *acc

			// Populate Links for this user
			userCopy.Links = m.getLinksForUser(acc.InternalID)

			// Populate ActiveSessions for this user (only active ones)
			userCopy.ActiveSessions = m.getActiveSessionsForUser(acc.InternalID)

			return &userCopy, nil
		}
	}
	return nil, nil
}

func (m *MockQueries) GetUserByLink(linkID uuid.UUID) (*auth.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	link, exists := m.links[linkID]
	if !exists {
		return nil, nil
	}

	for _, acc := range m.users {
		if acc.InternalID == link.UserID {
			// Create a copy of the user to avoid modifying the original
			userCopy := *acc

			// Populate Links for this user
			userCopy.Links = m.getLinksForUser(acc.InternalID)

			// Populate ActiveSessions for this user (only active ones)
			userCopy.ActiveSessions = m.getActiveSessionsForUser(acc.InternalID)

			return &userCopy, nil
		}
	}
	return nil, nil
}

// Application mock methods
func (m *MockQueries) InsertApplication(app *auth.Application) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	app.InternalID = len(m.applications) + 1
	m.applications[app.ID] = app
	return nil
}

func (m *MockQueries) UpdateApplication(app auth.Application) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	if _, exists := m.applications[app.ID]; !exists {
		return sql.ErrNoRows
	}

	m.applications[app.ID] = &app
	return nil
}

func (m *MockQueries) GetApplicationByID(id uuid.UUID) (*auth.Application, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	if app, exists := m.applications[id]; exists {
		return app, nil
	}
	return nil, nil
}

// Session mock methods
func (m *MockQueries) UpsertSessions(sessions ...auth.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	for _, session := range sessions {
		if session.ID == uuid.Nil {
			session.ID = uuid.New()
		}
		m.sessions[session.ID] = &session
	}
	return nil
}

// Link mock methods
func (m *MockQueries) UpsertLinks(links ...auth.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	for _, link := range links {
		if link.ID == uuid.Nil {
			link.ID = uuid.New()
		}
		m.links[link.ID] = &link
	}
	return nil
}

// Test helper methods
func (m *MockQueries) SetError(shouldError bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
	m.errorToReturn = err
}

func (m *MockQueries) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockQueries) AddUser(acc *auth.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if acc.ID == uuid.Nil {
		acc.ID = uuid.New()
	}
	m.users[acc.ID] = acc
}

func (m *MockQueries) AddApplication(app *auth.Application) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	m.applications[app.ID] = app
}

func (m *MockQueries) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[uuid.UUID]*auth.User)
	m.sessions = make(map[uuid.UUID]*auth.Session)
	m.applications = make(map[uuid.UUID]*auth.Application)
	m.links = make(map[uuid.UUID]*auth.Link)
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockQueries) GetAllUsers() []*auth.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*auth.User, 0, len(m.users))
	for _, acc := range m.users {
		users = append(users, acc)
	}
	return users
}

// Helper methods to populate Links and ActiveSessions
func (m *MockQueries) getLinksForUser(userInternalID int) []auth.Link {
	var links []auth.Link
	for _, link := range m.links {
		if link.UserID == userInternalID {
			links = append(links, *link)
		}
	}
	return links
}

func (m *MockQueries) getActiveSessionsForUser(userInternalID int) []auth.Session {
	var activeSessions []auth.Session
	for _, session := range m.sessions {
		if session.UserID == userInternalID && session.IsActive {
			activeSessions = append(activeSessions, *session)
		}
	}
	return activeSessions
}
