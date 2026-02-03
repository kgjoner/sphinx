package main

import (
	"log"
	"os"

	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/pkg/pgpool"
	"github.com/kgjoner/sphinx/test/testutils"
)

func main() {
	// Load configuration
	config.Must()

	// Connect to database
	pool, err := pgpool.New(config.Env.DATABASE_URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	action := "seed"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "seed":
		log.Println("Seeding test data...")
		seedData, err := testutils.SeedTestData(pool)
		if err != nil {
			log.Fatalf("Failed to seed test data: %v", err)
		}
		log.Printf("✅ Seeding complete!")
		log.Printf("   - Root App: %s", seedData.RootApp.ID)
		log.Printf("   - Test App: %s", seedData.TestApp.ID)
		log.Printf("   - Simple User: %s (email: %s)", seedData.SimpleUserID, seedData.SimpleUser.Email.String())
		log.Printf("   - Admin User: %s (email: %s)", seedData.AdminUserID, seedData.AdminUser.Email.String())

	case "clean":
		log.Println("Cleaning database...")
		if err := testutils.CleanDatabase(config.Env.DATABASE_URL); err != nil {
			log.Fatalf("Failed to clean database: %v", err)
		}
		log.Println("✅ Database cleaned!")

	default:
		log.Fatalf("Unknown action: %s. Use 'seed' or 'clean'", action)
	}
}
