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
	RootApplication.ID = uuid.MustParse(config.Env.ROOT_APP_ID)
}

func hashPassword(password string) string {
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes)
}

const RootAppSecret = "testsecret"

var RootApplication = &auth.Application{
	Name:      "Root Application",
	Secret:    hashPassword(RootAppSecret),
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

const CommonAppSecret = "commonappsecret"
const CommonRedirectUri = "http://common.app/callback"

var CommonApplication = &auth.Application{
	Name:                "Common Application",
	Secret:              hashPassword(CommonAppSecret),
	AllowedRedirectUris: []string{CommonRedirectUri},
	CreatedAt:           time.Now(),
	UpdatedAt:           time.Now(),
}

const AdminPassword = "AdminPassword123!"

var AdminAccount = &auth.Account{
	ID:                   uuid.New(),
	Email:                "admin@example.com",
	Password:             hashPassword(AdminPassword),
	Username:             "admin",
	IsActive:             true,
	HasEmailBeenVerified: true,
	CreatedAt:            time.Now(),
	UpdatedAt:            time.Now(),
}

var AdminRootLink = &auth.Link{
	ID:          uuid.New(),
	AccountID:   AdminAccount.InternalID,
	Application: *RootApplication,
	HasConsent:  true,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
	Roles:       []auth.Role{auth.RoleAdmin},
}

const SimpleUserPassword = "SimpleUserPassword123!"

var SimpleUserAccount = &auth.Account{
	ID:                   uuid.New(),
	Email:                "user@example.com",
	Password:             hashPassword(SimpleUserPassword),
	Username:             "user",
	IsActive:             true,
	HasEmailBeenVerified: true,
	CreatedAt:            time.Now(),
	UpdatedAt:            time.Now(),
}

var SimpleUserRootLink = &auth.Link{
	ID:          uuid.New(),
	AccountID:   SimpleUserAccount.InternalID,
	Application: *RootApplication,
	HasConsent:  true,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
}
