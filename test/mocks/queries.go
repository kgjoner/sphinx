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
	accounts     map[uuid.UUID]*auth.Account
	sessions     map[uuid.UUID]*auth.Session
	applications map[uuid.UUID]*auth.Application
	links        map[uuid.UUID]*auth.Link

	// Mock database errors for testing error scenarios
	shouldError   bool
	errorToReturn error
}

func NewMockQueries() *MockQueries {
	q := &MockQueries{
		accounts:     make(map[uuid.UUID]*auth.Account),
		sessions:     make(map[uuid.UUID]*auth.Session),
		applications: make(map[uuid.UUID]*auth.Application),
		links:        make(map[uuid.UUID]*auth.Link),
	}

	q.InsertApplication(RootApplication)
	AdminRootLink.Application = *RootApplication
	SimpleUserRootLink.Application = *RootApplication

	q.InsertApplication(CommonApplication)

	q.InsertAccount(AdminAccount)
	AdminRootLink.AccountId = AdminAccount.InternalId
	q.InsertAccount(SimpleUserAccount)
	SimpleUserRootLink.AccountId = SimpleUserAccount.InternalId

	q.UpsertLinks(*AdminRootLink, *SimpleUserRootLink)
	return q
}

// Account mock methods
func (m *MockQueries) InsertAccount(acc *auth.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	// Check for unique field violations
	if err := m.checkAccountUniqueConstraints(acc, nil); err != nil {
		return err
	}

	if acc.Id == uuid.Nil {
		acc.Id = uuid.New()
	}
	acc.InternalId = len(m.accounts) + 1
	m.accounts[acc.Id] = acc
	return nil
}

func (m *MockQueries) UpdateAccount(acc auth.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	if _, exists := m.accounts[acc.Id]; !exists {
		return sql.ErrNoRows
	}

	// Check for unique field violations (exclude the account being updated)
	if err := m.checkAccountUniqueConstraints(&acc, &acc.Id); err != nil {
		return err
	}

	m.accounts[acc.Id] = &acc
	return nil
}

// DuplicateKeyError represents a unique constraint violation
type DuplicateKeyError struct {
	Field string
	Value string
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate key value violates unique constraint on field account_%s_key: %s", e.Field, e.Value)
}

// checkAccountUniqueConstraints validates that the account doesn't violate unique field constraints
// excludeId allows skipping a specific account ID (useful for updates)
func (m *MockQueries) checkAccountUniqueConstraints(acc *auth.Account, excludeId *uuid.UUID) error {
	for _, existing := range m.accounts {
		// Skip the account being updated
		if excludeId != nil && existing.Id == *excludeId {
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

func (m *MockQueries) GetAccountById(id uuid.UUID) (*auth.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	if acc, exists := m.accounts[id]; exists {
		// Create a copy of the account to avoid modifying the original
		accountCopy := *acc

		// Populate Links for this account
		accountCopy.Links = m.getLinksForAccount(acc.InternalId)

		// Populate ActiveSessions for this account (only active ones)
		accountCopy.ActiveSessions = m.getActiveSessionsForAccount(acc.InternalId)

		return &accountCopy, nil
	}
	return nil, nil
}

func (m *MockQueries) GetAccountByEntry(entry string) (*auth.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	for _, acc := range m.accounts {
		if acc.Email.String() == entry || acc.Username == entry || acc.Document.String() == entry || acc.Phone.String() == entry {
			// Create a copy of the account to avoid modifying the original
			accountCopy := *acc

			// Populate Links for this account
			accountCopy.Links = m.getLinksForAccount(acc.InternalId)

			// Populate ActiveSessions for this account (only active ones)
			accountCopy.ActiveSessions = m.getActiveSessionsForAccount(acc.InternalId)

			return &accountCopy, nil
		}
	}
	return nil, nil
}

func (m *MockQueries) GetAccountByLink(linkId uuid.UUID) (*auth.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldError {
		return nil, m.errorToReturn
	}

	link, exists := m.links[linkId]
	if !exists {
		return nil, nil
	}

	for _, acc := range m.accounts {
		if acc.InternalId == link.AccountId {
			// Create a copy of the account to avoid modifying the original
			accountCopy := *acc
	
			// Populate Links for this account
			accountCopy.Links = m.getLinksForAccount(acc.InternalId)
	
			// Populate ActiveSessions for this account (only active ones)
			accountCopy.ActiveSessions = m.getActiveSessionsForAccount(acc.InternalId)
	
			return &accountCopy, nil
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

	if app.Id == uuid.Nil {
		app.Id = uuid.New()
	}
	app.InternalId = len(m.applications) + 1
	m.applications[app.Id] = app
	return nil
}

func (m *MockQueries) UpdateApplication(app auth.Application) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return m.errorToReturn
	}

	if _, exists := m.applications[app.Id]; !exists {
		return sql.ErrNoRows
	}

	m.applications[app.Id] = &app
	return nil
}

func (m *MockQueries) GetApplicationById(id uuid.UUID) (*auth.Application, error) {
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
		if session.Id == uuid.Nil {
			session.Id = uuid.New()
		}
		m.sessions[session.Id] = &session
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
		if link.Id == uuid.Nil {
			link.Id = uuid.New()
		}
		m.links[link.Id] = &link
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

func (m *MockQueries) AddAccount(acc *auth.Account) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if acc.Id == uuid.Nil {
		acc.Id = uuid.New()
	}
	m.accounts[acc.Id] = acc
}

func (m *MockQueries) AddApplication(app *auth.Application) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if app.Id == uuid.Nil {
		app.Id = uuid.New()
	}
	m.applications[app.Id] = app
}

func (m *MockQueries) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts = make(map[uuid.UUID]*auth.Account)
	m.sessions = make(map[uuid.UUID]*auth.Session)
	m.applications = make(map[uuid.UUID]*auth.Application)
	m.links = make(map[uuid.UUID]*auth.Link)
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockQueries) GetAllAccounts() []*auth.Account {
	m.mu.RLock()
	defer m.mu.RUnlock()

	accounts := make([]*auth.Account, 0, len(m.accounts))
	for _, acc := range m.accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}

// Helper methods to populate Links and ActiveSessions
func (m *MockQueries) getLinksForAccount(accountInternalId int) []auth.Link {
	var links []auth.Link
	for _, link := range m.links {
		if link.AccountId == accountInternalId {
			links = append(links, *link)
		}
	}
	return links
}

func (m *MockQueries) getActiveSessionsForAccount(accountInternalId int) []auth.Session {
	var activeSessions []auth.Session
	for _, session := range m.sessions {
		if session.AccountId == accountInternalId && session.IsActive {
			activeSessions = append(activeSessions, *session)
		}
	}
	return activeSessions
}
