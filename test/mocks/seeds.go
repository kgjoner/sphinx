package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	config.Must()
	config.Env.ROOT_APP_ID = RootApplication.ID.String()
}

func hashPassword(password string) string {
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes)
}

// ROOT APPLICATION
const RootAppSecret = "testsecret"

var RootApplication = &auth.Application{
	ID:                  uuid.MustParse("80cadd74-5ccd-41c4-9938-3c8961be04db"),
	Name:                "Root Application",
	Secret:              hashPassword(RootAppSecret),
	AllowedRedirectUris: []string{},
	PossibleRoles:       []auth.Role{auth.RoleAdmin, auth.RoleDev},
	CreatedAt:           time.Now(),
	UpdatedAt:           time.Now(),
}

// COMMON APPLICATION
const CommonAppSecret = "commonappsecret"
const CommonRedirectUri = "https://test.example.com/callback"

var CommonApplication = &auth.Application{
	ID:                  uuid.MustParse("abb00001-5ccd-41c4-9938-3c8961be04db"),
	Name:                "Test Application",
	Secret:              hashPassword(CommonAppSecret),
	AllowedRedirectUris: []string{CommonRedirectUri},
	PossibleRoles:       []auth.Role{auth.RoleAdmin},
	CreatedAt:           time.Now(),
	UpdatedAt:           time.Now(),
}

// ADMIN USER
const AdminPassword = "AdminPassword123!"

var AdminUser = &auth.User{
	ID:                   uuid.MustParse("abb00002-5ccd-41c4-9938-3c8961be04db"),
	Email:                "admin@example.com",
	Username:             "adminuser",
	Password:             hashPassword(AdminPassword),
	Phone:                "+5511988888888",
	Document:             "cpf:25576958071",
	IsActive:             true,
	HasEmailBeenVerified: true,
	HasPhoneBeenVerified: true,
	VerificationCodes:    map[auth.VerificationKind]string{},
	ExtraData: auth.ExtraData{
		Name:    "Admin",
		Surname: "User",
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

var AdminRootLink = &auth.Link{
	ID:          uuid.MustParse("abb00003-5ccd-41c4-9938-3c8961be04db"),
	UserID:      AdminUser.InternalID,
	Application: *RootApplication,
	HasConsent:  true,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
	Roles:       []auth.Role{auth.RoleAdmin},
}

// SIMPLE USER
const SimpleUserPassword = "SimpleUserPassword123!"

var SimpleUser = &auth.User{
	ID:                   uuid.MustParse("abb00004-5ccd-41c4-9938-3c8961be04db"),
	Email:                "simple@sphinx.test",
	Username:             "simpleuser",
	Password:             hashPassword(SimpleUserPassword),
	Phone:                "+5511999999999",
	Document:             "cpf:02496946031",
	IsActive:             true,
	HasEmailBeenVerified: true,
	HasPhoneBeenVerified: true,
	VerificationCodes:    map[auth.VerificationKind]string{},
	ExtraData: auth.ExtraData{
		Name:    "Simple",
		Surname: "User",
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

var SimpleUserRootLink = &auth.Link{
	ID:          uuid.MustParse("abb00005-5ccd-41c4-9938-3c8961be04db"),
	UserID:      SimpleUser.InternalID,
	Application: *RootApplication,
	HasConsent:  true,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
}
