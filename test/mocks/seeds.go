package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/pkg/security"
	"github.com/kgjoner/sphinx/internal/shared"
)

func init() {
	config.Must()
	config.Env.ROOT_APP_ID = RootApplication.ID.String()
}

var hasher = security.NewBcryptHasher()

// ROOT APPLICATION
const RootAppSecret = "TestS3cret"

var hashedRootAppSecret, _ = shared.NewHashedPassword(RootAppSecret, hasher)
var RootApplication = &access.Application{
	ID:                  uuid.MustParse("80cadd74-5ccd-41c4-9938-3c8961be04db"),
	Name:                "Root Application",
	Secret:              *hashedRootAppSecret,
	AllowedRedirectUris: []string{},
	PossibleRoles:       []access.Role{access.Admin, access.Manager},
	CreatedAt:           time.Now(),
	UpdatedAt:           time.Now(),
}

// COMMON APPLICATION
const CommonAppSecret = "commonAppS3cret"
const CommonRedirectUri = "https://test.example.com/callback"

var hashedCommonAppSecret, _ = shared.NewHashedPassword(CommonAppSecret, hasher)
var CommonApplication = &access.Application{
	ID:                  uuid.MustParse("abb00001-5ccd-41c4-9938-3c8961be04db"),
	Name:                "Test Application",
	Secret:              *hashedCommonAppSecret,
	AllowedRedirectUris: []string{CommonRedirectUri},
	PossibleRoles:       []access.Role{access.Admin},
	CreatedAt:           time.Now(),
	UpdatedAt:           time.Now(),
}

// ADMIN USER
const AdminPassword = "AdminPassword123!"

var hashedAdminPassword, _ = shared.NewHashedPassword(AdminPassword, hasher)
var AdminUser = &identity.User{
	ID:                   uuid.MustParse("abb00002-5ccd-41c4-9938-3c8961be04db"),
	Email:                "admin@example.com",
	Username:             "adminuser",
	Password:             *hashedAdminPassword,
	Phone:                "+5511988888888",
	Document:             "cpf:25576958071",
	IsActive:             true,
	HasEmailBeenVerified: true,
	HasPhoneBeenVerified: true,
	VerificationCodes:    map[identity.VerificationKind]string{},
	ExtraData: identity.ExtraData{
		Name:    "Admin",
		Surname: "User",
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

var AdminRootLink = &access.Link{
	ID:          uuid.MustParse("abb00003-5ccd-41c4-9938-3c8961be04db"),
	UserID:      AdminUser.ID,
	Application: *RootApplication,
	HasConsent:  true,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
	Roles:       []access.Role{access.Admin},
}

// SIMPLE USER
const SimpleUserPassword = "SimpleUserPassword123!"

var hashedSimpleUserPassword, _ = shared.NewHashedPassword(SimpleUserPassword, hasher)
var SimpleUser = &identity.User{
	ID:                   uuid.MustParse("abb00004-5ccd-41c4-9938-3c8961be04db"),
	Email:                "simple@sphinx.test",
	Username:             "simpleuser",
	Password:             *hashedSimpleUserPassword,
	Phone:                "+5511999999999",
	Document:             "cpf:02496946031",
	IsActive:             true,
	HasEmailBeenVerified: true,
	HasPhoneBeenVerified: true,
	VerificationCodes:    map[identity.VerificationKind]string{},
	ExtraData: identity.ExtraData{
		Name:    "Simple",
		Surname: "User",
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

var SimpleUserRootLink = &access.Link{
	ID:          uuid.MustParse("abb00005-5ccd-41c4-9938-3c8961be04db"),
	UserID:      SimpleUser.ID,
	Application: *RootApplication,
	HasConsent:  true,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
}
