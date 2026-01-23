package testutils

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	baserepo "github.com/kgjoner/sphinx/internal/repositories/base"
	"github.com/kgjoner/sphinx/test/mocks"
)

// SeedData contains commonly used test data
type SeedData struct {
	RootApp      auth.Application
	TestApp      auth.Application
	SimpleUser   auth.User
	SimpleUserID uuid.UUID
	AdminUser    auth.User
	AdminUserID  uuid.UUID
}

// SeedTestData seeds the database with common test data
func SeedTestData(pool *baserepo.Pool) (*SeedData, error) {
	ctx := context.Background()
	dao := pool.NewDAO(ctx)

	seed := &SeedData{}

	// Create Root Application
	err := dao.InsertApplication(mocks.RootApplication)
	if err != nil {
		return nil, fmt.Errorf("failed to seed root application: %v", err)
	}
	seed.RootApp = *mocks.RootApplication
	log.Printf("Seeded Root Application: %s", mocks.RootApplication.ID)

	// Create Test Application
	err = dao.InsertApplication(mocks.CommonApplication)
	if err != nil {
		return nil, fmt.Errorf("failed to seed test application: %v", err)
	}
	seed.TestApp = *mocks.CommonApplication
	log.Printf("Seeded Test Application: %s", mocks.CommonApplication.ID)

	// Create Simple User
	err = dao.InsertUser(mocks.SimpleUser)
	if err != nil {
		return nil, fmt.Errorf("failed to seed simple user: %v", err)
	}
	mocks.SimpleUserRootLink.Application = *mocks.RootApplication
	mocks.SimpleUserRootLink.UserID = mocks.SimpleUser.InternalID
	err = dao.UpsertLinks(*mocks.SimpleUserRootLink)
	if err != nil {
		return nil, fmt.Errorf("failed to seed simple user link: %v", err)
	}
	seed.SimpleUser = *mocks.SimpleUser
	seed.SimpleUserID = mocks.SimpleUser.ID
	log.Printf("Seeded Simple User: %s (email: %s, password: SimpleUserPassword123!)", mocks.SimpleUser.ID, mocks.SimpleUser.Email.String())

	// Create Admin User
	err = dao.InsertUser(mocks.AdminUser)
	if err != nil {
		return nil, fmt.Errorf("failed to seed admin user: %v", err)
	}
	mocks.AdminRootLink.Application = *mocks.RootApplication
	mocks.AdminRootLink.UserID = mocks.AdminUser.InternalID
	err = dao.UpsertLinks(*mocks.AdminRootLink)
	if err != nil {
		return nil, fmt.Errorf("failed to seed admin user link: %v", err)
	}
	seed.AdminUser = *mocks.AdminUser
	seed.AdminUserID = mocks.AdminUser.ID
	log.Printf("Seeded Admin User: %s (email: %s, password: AdminPassword123!)", mocks.AdminUser.ID, mocks.AdminUser.Email.String())

	return seed, nil
}

// GetSeedDataInfo returns the seed data structure with expected IDs and values
// This allows tests to know what data should exist without querying the database
func GetSeedDataInfo() *SeedData {
	return &SeedData{
		RootApp:    *mocks.RootApplication,
		TestApp:    *mocks.CommonApplication,
		SimpleUser: *mocks.SimpleUser,
		AdminUser:  *mocks.AdminUser,
	}
}
