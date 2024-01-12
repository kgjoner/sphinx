package auth

import (
	"sort"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func (a *Account) InitSession(f *SessionCreationFields) (*Session, *Jwt, *Jwt, error) {
	// Check if link exists
	isLinked := false
	for _, l := range a.Links {
		if l.Application.Id == f.Application.Id {
			isLinked = true
			break
		}
	}

	if !isLinked {
		return nil, nil, nil, normalizederr.NewRequestError("Account is not linked to desired application.", "")
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
				return nil, nil, nil, err
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
	refreshToken, err := newJwt(*a, s.Id, "refresh")
	if err != nil {
		return nil, nil, nil, err
	}

	s.updateRefreshToken(*refreshToken)

	accessToken, err := newJwt(*a, s.Id)
	if err != nil {
		return nil, nil, nil, err
	}

	a.ActiveSessions = append(a.ActiveSessions, *s)
	return s, accessToken, refreshToken, validator.Validate(s)
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

func (s *Session) updateRefreshToken(token Jwt) {
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

/* ==============================================================================
	JWT
============================================================================== */

type Jwt struct {
	jwt.Token
	Claims       jwtClaims
	signedString string
}

func ParseJwtString(str string) (*Jwt, error) {
	token, err := jwt.Parse(str, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Environment.JWT.SECRET), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, normalizederr.NewRequestError("Invalid token", "")
	}

	claims, ok := token.Claims.(jwtClaims)
	if !ok {
		return nil, normalizederr.NewRequestError("Badly formatted token", "")
	}

	return &Jwt{*token, claims, str}, nil
}

func newJwt(a Account, sessionId uuid.UUID, kindFlag ...string) (*Jwt, error) {
	s := a.Session(sessionId)
	if s == nil {
		return nil, normalizederr.NewRequestError("Account and session do not match.", "")
	}

	now := time.Now()
	var kind string
	var duration time.Duration
	if len(kindFlag) >= 1 && kindFlag[0] == "refresh" {
		kind = "refresh"
		duration = time.Second * time.Duration(config.Environment.JWT.REFRESH_LIFE_TIME_IN_SEC)
	} else {
		kind = "access"
		duration = time.Second * time.Duration(config.Environment.JWT.ACCESS_LIFE_TIME_IN_SEC)
	}

	claims := jwtClaims{
		a.Id,
		s.Application.Id,
		now,
		now.Add(duration),
		s.Id,
		kind,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenAsSignedString, err := token.SignedString([]byte(config.Environment.JWT.SECRET))
	if err != nil {
		return nil, err
	}

	return &Jwt{*token, claims, tokenAsSignedString}, nil
}

func (t Jwt) IsExpired() bool {
	now := time.Now()
	return t.Claims.Exp.Before(now)
}

func (t Jwt) String() string {
	return t.signedString
}

type jwtClaims struct {
	Sub       uuid.UUID `json:"sub"`
	Aud       uuid.UUID `json:"aud"`
	Iat       time.Time `json:"iat"`
	Exp       time.Time `json:"exp"`
	SessionId uuid.UUID `json:"sessionId"`
	Kind      string    `json:"kind" validate:"oneof=refresh access"`
}

func (c jwtClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Aud.String()}, nil
}
func (c jwtClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: c.Exp}, nil
}
func (c jwtClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: c.Iat}, nil
}
func (c jwtClaims) GetIssuer() (string, error) {
	return "", nil
}
func (c jwtClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{}, nil
}
func (c jwtClaims) GetSubject() (string, error) {
	return c.Sub.String(), nil
}
