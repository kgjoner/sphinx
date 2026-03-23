package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/access/accessrepo"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/domains/identity/identrepo"
	"github.com/kgjoner/sphinx/internal/pkg/pgpool"
	"github.com/kgjoner/sphinx/test/mocks"
)

// SeedData contains commonly used test data
type SeedData struct {
	RootApp        access.Application
	TestApp        access.Application
	SimpleUser     identity.User
	SimpleUserID   uuid.UUID
	SimpleRootLink access.Link
	AdminUser      identity.User
	AdminUserID    uuid.UUID
	AdminRootLink  access.Link
}

// SeedTestData seeds the database with common test data
func SeedTestData(pool *pgpool.Pool) (*SeedData, error) {
	ctx := context.Background()
	seed := &SeedData{}

	_, err := pool.WithTx(ctx, nil, func(tx *sql.Tx) (any, error) {
		identRepo := identrepo.NewFactory().NewDAO(ctx, tx)
		accessRepo := accessrepo.NewFactory().NewDAO(ctx, tx)

		// Create Root Application
		err := accessRepo.InsertApplication(mocks.RootApplication)
		if err != nil {
			return nil, fmt.Errorf("failed to seed root application: %v", err)
		}
		seed.RootApp = *mocks.RootApplication
		log.Printf("Seeded Root Application: %s", mocks.RootApplication.ID)

		// Create Test Application
		err = accessRepo.InsertApplication(mocks.CommonApplication)
		if err != nil {
			return nil, fmt.Errorf("failed to seed test application: %v", err)
		}
		seed.TestApp = *mocks.CommonApplication
		log.Printf("Seeded Test Application: %s", mocks.CommonApplication.ID)

		// Create Simple User
		err = identRepo.InsertUser(mocks.SimpleUser)
		if err != nil {
			return nil, fmt.Errorf("failed to seed simple user: %v", err)
		}
		mocks.SimpleUserRootLink.Application = *mocks.RootApplication
		err = accessRepo.InsertLink(mocks.SimpleUserRootLink)
		if err != nil {
			return nil, fmt.Errorf("failed to seed simple user link: %v", err)
		}
		seed.SimpleUser = *mocks.SimpleUser
		seed.SimpleUserID = mocks.SimpleUser.ID
		seed.SimpleRootLink = *mocks.SimpleUserRootLink
		log.Printf("Seeded Simple User: %s (email: %s, password: SimpleUserPassword123!)", mocks.SimpleUser.ID, mocks.SimpleUser.Email.String())

		// Create Admin User
		err = identRepo.InsertUser(mocks.AdminUser)
		if err != nil {
			return nil, fmt.Errorf("failed to seed admin user: %v", err)
		}
		mocks.AdminRootLink.Application = *mocks.RootApplication

		err = accessRepo.InsertLink(mocks.AdminRootLink)
		if err != nil {
			return nil, fmt.Errorf("failed to seed admin user link: %v", err)
		}
		seed.AdminUser = *mocks.AdminUser
		seed.AdminUserID = mocks.AdminUser.ID
		seed.AdminRootLink = *mocks.AdminRootLink
		log.Printf("Seeded Admin User: %s (email: %s, password: AdminPassword123!)", mocks.AdminUser.ID, mocks.AdminUser.Email.String())

		return seed, nil
	})

	return seed, err
}

// GetSeedDataInfo returns the seed data structure with expected IDs and values
// This allows tests to know what data should exist without querying the database
func GetSeedDataInfo() *SeedData {
	return &SeedData{
		RootApp:        *mocks.RootApplication,
		TestApp:        *mocks.CommonApplication,
		SimpleUser:     *mocks.SimpleUser,
		SimpleUserID:   mocks.SimpleUser.ID,
		SimpleRootLink: *mocks.SimpleUserRootLink,
		AdminUser:      *mocks.AdminUser,
		AdminUserID:    mocks.AdminUser.ID,
		AdminRootLink:  *mocks.AdminRootLink,
	}
}
