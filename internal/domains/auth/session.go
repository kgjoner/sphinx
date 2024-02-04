package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	InternalId                     int             `json:"-"`
	Id                             uuid.UUID       `json:"id" validator:"required"`
	AccountId                      int             `json:"account_id" validator:"required"`
	Application                    Application     `json:"application" validator:"required"`
	RefreshToken                   string          `json:"refresh_token" validator:"required"`
	RefreshedAt                    htypes.NullTime `json:"refreshed_at"`
	ElapsedMinutesBetweenRefreshes []int           `json:"elapsed_minutes_between_refreshes"`
	RefreshesCount                 int             `json:"refreshes_count"`
	Device                         string          `json:"device" validator:"required"`
	Ip                             string          `json:"ip" validator:"required"`
	IsActive                       bool            `json:"is_active"`
	TerminatedAt                   htypes.NullTime `json:"terminated_at"`
	CreatedAt                      time.Time       `json:"created_at" validator:"required"`
	UpdatedAt                      time.Time       `json:"updated_at" validator:"required"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type SessionCreationFields struct {
	Application Application `json:"application" validator:"required"`
	Device      string      `json:"device" validator:"required"`
	Ip          string      `json:"ip" validator:"required"`
}

func newSession(acc Account, f *SessionCreationFields) *Session {
	now := time.Now()
	s := &Session{
		Id:                             uuid.New(),
		AccountId:                      acc.InternalId,
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
	s.RefreshToken = ""
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
