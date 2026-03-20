package sphinx

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/apperr"
	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/cornucopia/v3/httpclient"
	"github.com/kgjoner/cornucopia/v3/structop"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

/* ==============================================================================
	Legacy Support
============================================================================== */

// Deprecated: use Client instead
type Service struct {
	httpApi   *httpclient.HTTPUtil
	appID     string
	appSecret string
	appToken  string
}

// Deprecated: use NewClient instead
func New(baseURL, appID, appSecret string) *Service {
	httpApi := httpclient.New(baseURL)
	appToken := base64.StdEncoding.EncodeToString([]byte(appID + ":" + appSecret))

	svc := &Service{
		httpApi:   httpApi,
		appID:     appID,
		appSecret: appSecret,
		appToken:  appToken,
	}

	return svc
}

/* ==============================================================================
	User
============================================================================== */

// Deprecated: use UserView or Subject instead
type User struct {
	ID       uuid.UUID          `json:"id" validate:"required"`
	Email    prim.Email       `json:"email" validate:"required"`
	Phone    prim.PhoneNumber `json:"phone,omitempty"`
	Username string             `json:"username,omitempty" validate:"wordID"`
	Document prim.Document    `json:"document,omitempty"`
	Name     string             `json:"name,omitempty"`
	Surname  string             `json:"surname,omitempty"`
	Address  prim.Address     `json:"address,omitempty"`

	IsActive             bool               `json:"isActive"`
	PendingEmail         prim.Email       `json:"pendingEmail,omitempty"`
	HasEmailBeenVerified bool               `json:"hasEmailBeenVerified"`
	PendingPhone         prim.PhoneNumber `json:"pendingPhone,omitempty"`
	HasPhoneBeenVerified bool               `json:"hasPhoneBeenVerified"`
	UsernameUpdatedAt    prim.NullTime    `json:"usernameUpdatedAt"`
	Link                 *access.LinkView   `json:"link,omitempty"`
}

func (a User) DisplayName() string {
	if a.Name != "" {
		return a.Name
	}

	email := a.Email.String()
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at]
	}
	return email
}

func (a User) IsAdmin() bool {
	for _, r := range a.Link.Roles {
		if r == access.Admin {
			return true
		}
	}

	return false
}

func (a User) IsDev() bool {
	for _, r := range a.Link.Roles {
		if r == "DEV" {
			return true
		}
	}

	return false
}

func (a User) HasRole(role string) bool {
	for _, r := range a.Link.Roles {
		if string(r) == role {
			return true
		}
	}

	return false
}

/* ==============================================================================
	Methods
============================================================================== */

// v1.7+ endpoint does not return link along User anymore. Retrieve it separately.
func (s Service) getLink(signedToken string) (*access.LinkView, error) {
	token, _ := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		return []byte("invalid_secret"), nil // not checking validity here
	})

	var claims map[string]string
	ms, _ := json.Marshal(token.Claims)
	_ = json.Unmarshal(ms, &claims)

	userID := claims["sub"]
	appID := claims["aud"]

	var respData httpserver.Success[access.LinkView]
	_, err := s.httpApi.Get("/user/"+userID+"/link/"+appID, &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + signedToken,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

// Get token owner's data.
func (s Service) Me(token string) (*User, error) {
	var respData httpserver.Success[User]
	_, err := s.httpApi.Get("/user/me", &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	link, err := s.getLink(token)
	if err != nil {
		return nil, err
	}

	respData.Data.Link = link
	return &respData.Data, nil
}

// Get target user's data. Return error if target user does not exist.
//
// Token owner must be an admin.
func (s Service) User(userID uuid.UUID, token string) (*User, error) {
	var respData httpserver.Success[User]
	_, err := s.httpApi.Get("/user/"+userID.String(), &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	link, err := s.getLink(token)
	if err != nil {
		return nil, err
	}

	respData.Data.Link = link
	return &respData.Data, nil
}

// Get target user's email. Return error if target user does not exist.
func (s Service) EmailOf(userID uuid.UUID) (prim.Email, error) {
	var respData httpserver.Success[prim.Email]
	_, err := s.httpApi.Get("/user/"+userID.String()+"/email", &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(&respData)

	if err != nil {
		return "", err
	}

	return respData.Data, nil
}

// Create a simple user for the informed email.
func (s Service) NewUser(email prim.Email, password string) (userID uuid.UUID, err error) {
	body := map[string]any{
		"email":    email,
		"password": password,
	}

	var respData httpserver.Success[User]
	_, err = s.httpApi.Post("/user", body, nil)(&respData)

	if err != nil {
		return uuid.Nil, err
	}

	return respData.Data.ID, nil
}

// Check whether entry exists.
func (s Service) DoesEntryExist(entry string) (bool, error) {
	var respData httpserver.Success[bool]
	_, err := s.httpApi.Get("/user/existence", &httpclient.Options{
		Headers: map[string]string{
			"X-Entry": entry,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

// Get user id by their entry. Return zero value if entry is not found.
func (s Service) UserIDByEntry(entry string) (uuid.UUID, error) {
	var respData httpserver.Success[uuid.UUID]
	_, err := s.httpApi.Get("/user/id", &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
			"X-Entry":       entry,
		},
	})(&respData)

	if err != nil {
		return uuid.Nil, err
	}

	return respData.Data, nil
}

// Add roles and/or grantings to target user.
func (s Service) GrantPermissions(userID uuid.UUID, roles []string) (bool, error) {
	body := map[string]any{
		"shouldRemove": sql.NullBool{
			Valid: true,
			Bool:  false,
		},
	}

	if len(roles) > 0 {
		body["roles"] = roles
	}

	var respData httpserver.Success[bool]
	_, err := s.httpApi.Patch("/user/"+userID.String()+"/permission", body, &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

// Remove roles and/or grantings from target user.
func (s Service) RevokePermissions(userID uuid.UUID, roles []string) (bool, error) {
	body := map[string]any{
		"shouldRemove": sql.NullBool{
			Valid: true,
			Bool:  true,
		},
	}

	if len(roles) > 0 {
		body["roles"] = roles
	}

	var respData httpserver.Success[bool]
	_, err := s.httpApi.Patch("/user/"+userID.String()+"/permission", body, &httpclient.Options{
		Headers: map[string]string{
			"Authorization": "Basic " + s.appToken,
		},
	})(&respData)

	if err != nil {
		return false, err
	}

	return respData.Data, nil
}

type LoginOutput struct {
	UserID       uuid.UUID `json:"userID"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"`
}

// Deprecated: use ExternalLoginBody instead if using sphinx.Client instead of sphinx.Service
type ExternalAuthBody struct {
	ProviderName    string `validate:"required"`
	Params          map[string]string
	Body            map[string]string
	ConsentRelation bool
	ConsentCreation bool
	Email           string
}

func (s Service) ExternalAuth(authorization string, body ExternalAuthBody, clientIP string, userAgent string, languages ...string) (*LoginOutput, error) {
	mapBody := structop.New(body).Map()

	var respData httpserver.Success[LoginOutput]
	_, err := s.httpApi.Post("/auth/external", mapBody, &httpclient.Options{
		Headers: map[string]string{
			"Authorization":   authorization,
			"Accept-Language": strings.Join(languages, ","),
			"X-Forwarded-For": clientIP,
			"User-Agent":      userAgent,
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

/* ==============================================================================
	Middlewares
============================================================================== */

type ctxKey string

// Deprecated: use Middlewares instead and define your own
// context values with Authorizer.AuthorizeSubject.
const (
	ActorCtxKey ctxKey = "sphinx_actor"
	TokenCtxKey ctxKey = "sphinx_token"
)

// Deprecated: use middlewares instead; created either with NewMiddlewares or
// Client.Middlewares
type Middlewares struct {
	sphinx *Service
}

func (s *Service) Middlewares() *Middlewares {
	return &Middlewares{
		sphinx: s,
	}
}

// Ensure authentication via bearer token
func (m Middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" || authHeaderParts[1] == "" {
			httpserver.HTTPError(auth.ErrInvalidAccess, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		user, err := m.sphinx.Me(tokenStr)
		if err != nil {
			httpserver.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, *user)
		ctx = context.WithValue(ctx, TokenCtxKey, tokenStr)
		ctx = context.WithValue(ctx, httpserver.ActorLogKey, user.ID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// If authorization header is present, ensure authentication via bearer token. Otherwise, allow request forward.
func (m Middlewares) TryAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		m.Authenticate(next).ServeHTTP(w, r)
	})
}

// Ensure authenticated user has at least one of listed roles. Admin users are always allowed.
func (m Middlewares) Guard(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actorValue := r.Context().Value(ActorCtxKey)
			if actorValue == nil {
				httpserver.HTTPError(auth.ErrInvalidAccess, w, r)
				return
			}

			actor := actorValue.(User)
			if actor.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			for _, p := range roles {
				if actor.HasRole(p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			err := apperr.NewForbiddenError("user does not have enough permission")
			httpserver.HTTPError(err, w, r)
		})
	}
}
