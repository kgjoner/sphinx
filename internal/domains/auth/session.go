package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Session struct {
	ID           uuid.UUID    `validate:"required"`
	SubjectID    uuid.UUID    `validate:"required"`
	SubjectEmail htypes.Email `validate:"required"`
	SubjectName  string       `validate:"required"`

	AudienceID uuid.UUID `validate:"required"`
	Roles      []string

	IP                             string `validate:"required"`
	Device                         string `validate:"required"`
	RefreshToken                   shared.HashedData
	RefreshedAt                    htypes.NullTime
	ElapsedMinutesBetweenRefreshes []int
	RefreshesCount                 int

	IsActive        bool
	isAuthenticated bool

	TerminatedAt htypes.NullTime
	CreatedAt    time.Time `validate:"required"`
	UpdatedAt    time.Time `validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type SessionCreationFields struct {
	Device string
	IP     string
}

func NewSession(f SessionCreationFields, p Principal, proof shared.AuthProof) (*Session, error) {
	switch proofTyped := proof.(type) {
	case *shared.PasswordProof:
		if !proofTyped.ValidFor(p.Password) {
			return nil, shared.ErrInvalidCredentials
		}
	case *GrantProof:
		if !proofTyped.ValidFor(p) {
			return nil, shared.ErrInvalidCredentials
		}
	case *ExternalLoginProof:
		if !proofTyped.ValidFor(p) {
			return nil, shared.ErrInvalidCredentials
		}
	default:
		return nil, shared.ErrInvalidProof
	}

	now := time.Now()
	s := &Session{
		ID:                             uuid.New(),
		SubjectID:                      p.ID,
		SubjectEmail:                   p.Email,
		SubjectName:                    p.Name,
		AudienceID:                     p.AudienceID,
		Roles:                          p.Roles,
		Device:                         f.Device,
		IP:                             f.IP,
		IsActive:                       true,
		CreatedAt:                      now,
		UpdatedAt:                      now,
		ElapsedMinutesBetweenRefreshes: []int{},
		isAuthenticated:                true,
	}

	return s, validator.Validate(s)
}

/* ==============================================================================
	METHODS
============================================================================== */

// Validate checks if the session is still active and matches the provided actor's details.
func (s *Session) Validate(actor *shared.Actor) error {
	if !s.IsActive || actor.SessionID != s.ID || actor.ID != s.SubjectID || actor.AudienceID != s.AudienceID {
		return ErrInvalidSession
	}

	return nil
}

func (s *Session) Authenticate(proof shared.AuthProof) error {
	switch proofTyped := proof.(type) {
	case *shared.DataProof:
		if !proofTyped.ValidFor(s.RefreshToken) {
			return shared.ErrInvalidCredentials
		}
	default:
		return shared.ErrInvalidProof
	}

	s.isAuthenticated = true
	return nil
}

func (s *Session) Terminate() error {
	now := time.Now()
	s.IsActive = false
	s.TerminatedAt = htypes.NullTime{Time: now}
	s.UpdatedAt = now
	return validator.Validate(s)
}

func (s *Session) UpdateRefreshToken(token shared.HashedData) {
	// First time setting the refresh token
	if s.RefreshToken.IsZero() {
		s.RefreshToken = token
		return
	}

	if len(s.ElapsedMinutesBetweenRefreshes) >= 1000 {
		sum := 0
		for _, minutes := range s.ElapsedMinutesBetweenRefreshes {
			sum += minutes
		}

		s.ElapsedMinutesBetweenRefreshes = []int{sum / len(s.ElapsedMinutesBetweenRefreshes)}
	}

	now := time.Now()
	lastRefreshTime := s.RefreshedAt.Time
	if lastRefreshTime.IsZero() {
		lastRefreshTime = s.CreatedAt
	}
	elapsedTime := now.Sub(lastRefreshTime)

	s.RefreshToken = token
	s.ElapsedMinutesBetweenRefreshes = append(s.ElapsedMinutesBetweenRefreshes, int(elapsedTime.Minutes()))
	s.RefreshesCount += 1
	s.RefreshedAt = htypes.NullTime{Time: now}
	s.UpdatedAt = now
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

/* ==============================================================================
	VIEWS
============================================================================== */

type SessionView struct {
	ID                             uuid.UUID       `json:"id"`
	SubjectID                      uuid.UUID       `json:"subjectId"`
	SubjectEmail                   htypes.Email    `json:"subjectEmail"`
	AudienceID                     uuid.UUID       `json:"audienceId"`
	RefreshedAt                    htypes.NullTime `json:"refreshedAt"`
	ElapsedMinutesBetweenRefreshes []int           `json:"elapsedMinutesBetweenRefreshes"`
	RefreshesCount                 int             `json:"refreshesCount"`
	Device                         string          `json:"device"`
	IP                             string          `json:"ip"`
	IsActive                       bool            `json:"isActive"`
	TerminatedAt                   htypes.NullTime `json:"terminatedAt"`
	CreatedAt                      time.Time       `json:"createdAt"`
	UpdatedAt                      time.Time       `json:"updatedAt"`
}

func (s Session) View() SessionView {
	return SessionView{
		ID:                             s.ID,
		SubjectID:                      s.SubjectID,
		SubjectEmail:                   s.SubjectEmail,
		AudienceID:                     s.AudienceID,
		RefreshedAt:                    s.RefreshedAt,
		ElapsedMinutesBetweenRefreshes: s.ElapsedMinutesBetweenRefreshes,
		RefreshesCount:                 s.RefreshesCount,
		Device:                         s.Device,
		IP:                             s.IP,
		IsActive:                       s.IsActive,
		TerminatedAt:                   s.TerminatedAt,
		CreatedAt:                      s.CreatedAt,
		UpdatedAt:                      s.UpdatedAt,
	}
}

func (s Session) ToSubject() (*Subject, error) {
	if !s.isAuthenticated {
		return nil, ErrInvalidSession
	}

	roles := s.Roles
	if roles == nil {
		roles = []string{}
	}

	return &Subject{
		Kind:       shared.KindUser,
		ID:         s.SubjectID,
		Email:      s.SubjectEmail,
		Name:       s.SubjectName,
		AudienceID: s.AudienceID,
		Roles:      roles,
		SessionID:  s.ID,
	}, nil
}
