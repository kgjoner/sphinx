package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
	"github.com/kgjoner/cornucopia/utils/structop"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type Application struct {
	InternalId int       `json:"-"`
	Id         uuid.UUID `json:"id" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Grantings  []string  `json:"grantings"`

	Secret              string   `json:"-" validate:"required"`
	AllowedRedirectUris []string `json:"allowedRedirectUris" validate:"uri"`
	Brand               brand    `json:"brand"`

	HasValidCredentials bool `json:"-"`

	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

type brand struct {
	LogoUrl         string `json:"logoUrl" validate:"uri"`
	StyleUrl        string `json:"styleUrl" validate:"uri"`
	IsValidOnEmail  bool   `json:"isValidOnEmail"`
	IsValidOnClient bool   `json:"isValidOnClient"`
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type ApplicationCreationFields struct {
	Name                string   `json:"name" validate:"required"`
	Grantings           []string `json:"grantings"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
	Brand               brand    `json:"brand"`
}

func NewApplication(f *ApplicationCreationFields, actor Account) (app *Application, secret string, err error) {
	actorApp := actor.AuthedSession.Application
	if !actorApp.isRoot() || !(actor.HasRole(actorApp, RoleValues.ADMIN) || actor.HasGranting(actorApp, "DEV")) {
		return nil, "", normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	secret = generateAppSecret()
	now := time.Now()
	created := &Application{
		Id:        uuid.New(),
		Name:      f.Name,
		Grantings: f.Grantings,

		Secret:              hashPassword(secret),
		AllowedRedirectUris: f.AllowedRedirectUris,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, secret, validator.Validate(created)
}

func generateAppSecret() string {
	return pwdgen.Generate(42, "lower", "upper", "number")
}

/* ==============================================================================
	METHODS
============================================================================== */

func (a Application) IsValid() error {
	errs := make(map[string]error)

	roles := structop.New(RoleValues).Map()
	for _, g := range a.Grantings {
		for _, r := range roles {
			if strings.ToUpper(g) == r {
				key := fmt.Sprintf("Grantings[%v]", g)
				errs[key] = fmt.Errorf("%v is a reserved role, it cannot be a granting", g)
				break
			}
		}
	}

	if len(errs) != 0 {
		return normalizederr.NewValidationErrorFromMap(errs)
	}

	return nil
}

type ApplicationEditableFields struct {
	Name                string   `json:"name"`
	Grantings           []string `json:"grantings"`
	AllowedRedirectUris []string `json:"allowedRedirectUris"`
	Brand               brand    `json:"brand"`
}

func (a *Application) Edit(f *ApplicationEditableFields, actor Account) error {
	actorApp := actor.AuthedSession.Application
	if actorApp.Id != a.Id || !actor.HasRoleOnAuth(RoleValues.ADMIN) {
		return normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	structop.New(a).Update(f)
	a.UpdatedAt = time.Now()
	return validator.Validate(a)
}

func (a *Application) GenerateNewSecret(actor Account) (secret string, err error) {
	actorApp := actor.AuthedSession.Application
	if actorApp.Id != a.Id || !actor.HasRoleOnAuth(RoleValues.ADMIN) {
		return "", normalizederr.NewForbiddenError("Does not have permission to execute this action.")
	}

	secret = generateAppSecret()
	a.Secret = hashPassword(secret)
	a.UpdatedAt = time.Now()
	return secret, validator.Validate(a)
}

func (a Application) isRoot() bool {
	return a.Id.String() == config.Env.ROOT_APP_ID
}

func (a *Application) DoesSecretMatch(secret string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Secret), []byte(secret))
	return err == nil
}

func (a *Application) Authenticate(secret string) error {
	if !a.DoesSecretMatch(secret) {
		return normalizederr.NewFatalUnauthorizedError("invalid credentials", errcode.InvalidAccess)
	}

	a.HasValidCredentials = true
	return nil
}

func (a Application) IsAuthenticated() bool {
	return a.HasValidCredentials
}
