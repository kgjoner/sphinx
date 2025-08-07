package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/validator"
)

type Session struct {
	InternalID                     int             `json:"-"`
	ID                             uuid.UUID       `json:"id" validate:"required"`
	UserID                         int             `json:"userID" validate:"required"`
	Application                    Application     `json:"application" validate:"required"`
	RefreshToken                   string          `json:"refreshToken" validate:"required"`
	RefreshedAt                    htypes.NullTime `json:"refreshedAt"`
	ElapsedMinutesBetweenRefreshes []int           `json:"elapsedMinutesBetweenRefreshes"`
	RefreshesCount                 int             `json:"refreshesCount"`
	Device                         string          `json:"device" validate:"required"`
	Ip                             string          `json:"ip"`
	IsActive                       bool            `json:"isActive"`
	TerminatedAt                   htypes.NullTime `json:"terminatedAt"`
	CreatedAt                      time.Time       `json:"createdAt" validate:"required"`
	UpdatedAt                      time.Time       `json:"updatedAt" validate:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type SessionCreationFields struct {
	Application Application `json:"application" validate:"required"`
	Device      string      `json:"device" validate:"required"`
	Ip          string      `json:"ip" validate:"required"`
}

func newSession(user User, f *SessionCreationFields) *Session {
	now := time.Now()
	s := &Session{
		ID:                             uuid.New(),
		UserID:                         user.InternalID,
		Application:                    f.Application,
		Device:                         f.Device,
		Ip:                             f.Ip,
		IsActive:                       true,
		CreatedAt:                      now,
		UpdatedAt:                      now,
		ElapsedMinutesBetweenRefreshes: []int{},
	}

	return s
}

/* ==============================================================================
	METHODS
============================================================================== */

func (s *Session) terminate() error {
	now := time.Now()
	s.IsActive = false
	s.TerminatedAt = htypes.NullTime{Time: now}
	s.UpdatedAt = now
	return validator.Validate(s)
}

func (s *Session) updateRefreshToken(token authToken) {
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

	s.RefreshToken = hashData(token.String())
	s.ElapsedMinutesBetweenRefreshes = append(s.ElapsedMinutesBetweenRefreshes, int(elapsedTime.Minutes()))
	s.RefreshesCount += 1
	s.RefreshedAt = htypes.NullTime{Time: now}
	s.UpdatedAt = now
}

func (s Session) doesRefreshTokenMatch(signedString string) bool {
	return s.RefreshToken == hashData(signedString)
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
