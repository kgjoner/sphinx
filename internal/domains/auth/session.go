package auth

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/sphinx/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	InternalId   int             `json:"-"`
	Id           uuid.UUID       `json:"id" validator:"required"`
	AccountId    int             `json:"-" validator:"required"`
	Application  Application     `json:"-" validator:"required"`
	RefreshToken string          `json:"-" validator:"required"`
	Device       string          `json:"device" validator:"required"`
	Ip           string          `json:"ip" validator:"required"`
	IsActive     bool            `json:"-"`
	TerminatedAt htypes.NullTime `json:"terminatedAt"`
	CreatedAt    time.Time       `json:"createdAt" validator:"required"`
	UpdatedAt    time.Time       `json:"updatedAt" validator:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type SessionCreationFields struct {
	Application Application `json:"-" validator:"required"`
	Device      string      `json:"device" validator:"required"`
	Ip          string      `json:"ip" validator:"required"`
}

func (a *Account) InitSession(f *SessionCreationFields) (*authToken, *authToken, error) {
	// Check if link exists
	isLinked := false
	for _, l := range a.Links {
		if l.Application.Id == f.Application.Id {
			isLinked = true
			break
		}
	}

	if !isLinked {
		return nil, nil, normalizederr.NewRequestError("Account is not linked to desired application.", "")
	}

	// Terminate exceeding sessions
	if concurrentSessions, maxConcurrentSessions := SessionSortableByAge(a.Sessions(f.Application)),
		config.Environment.MAX_CONCURRENT_SESSIONS; maxConcurrentSessions > 0 && len(concurrentSessions) >= maxConcurrentSessions {
		previousExcess := len(concurrentSessions) - maxConcurrentSessions
		/* If all goes well, previousExcess will be zero. There will be only one session to terminate for preventing
		new session to exceed max limit */
		sort.Sort(concurrentSessions)
		sessionsToTerminate := concurrentSessions[:(previousExcess + 1)]
		for _, s := range sessionsToTerminate {
			_, err := a.TerminateSession(s.Id)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	// Create session
	now := time.Now()
	s := &Session{
		Id:          uuid.New(),
		AccountId:   a.InternalId,
		Application: f.Application,
		Device:      f.Device,
		Ip:          f.Ip,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Generate tokens
	refreshToken, err := newAuthToken(*a, s.Id, "refresh")
	if err != nil {
		return nil, nil, err
	}

	s.updateRefreshToken(*refreshToken)

	accessToken, err := newAuthToken(*a, s.Id)
	if err != nil {
		return nil, nil, err
	}

	a.ActiveSessions = append(a.ActiveSessions, *s)
	return accessToken, refreshToken, validator.Validate(s)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (s *Session) terminate() error {
	now := time.Now()
	s.IsActive = false
	s.RefreshToken = ""
	s.TerminatedAt = htypes.NullTime{Time: now}
	s.UpdatedAt = now
	return validator.Validate(s)
}

func (s *Session) updateRefreshToken(token authToken) {
	s.RefreshToken = hashData(token.String())
	s.UpdatedAt = time.Now()
}

func (s Session) doesRefreshTokenMatch(signedString string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s.RefreshToken), []byte(signedString))
	return err == nil
}

/* ==============================================================================
	SORT
============================================================================== */

type SessionSortableByAge []Session

func (a SessionSortableByAge) Len() int {
	return len(a)
}

func (a SessionSortableByAge) Less(i, j int) bool {
	return a[i].CreatedAt.Before(a[j].CreatedAt)
}

func (a SessionSortableByAge) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
