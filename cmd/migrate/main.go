package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/kgjoner/sphinx/internal/repositories/base/migrations"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	migrationFiles := migrations.Files
	// 1. Load migrations from embedded filesystem
	d, err := iofs.New(migrationFiles, ".")
	if err != nil {
		log.Fatalf("failed to create iofs driver: %v", err)
	}

	// 2. Initialize the migrator
	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		log.Fatalf("failed to initialize migrator: %v", err)
	}

	// 3. Execute "Up" migrations
	fmt.Println("Running migrations...")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No new migrations to apply.")
		} else {
			log.Fatalf("failed to run migrations: %v", err)
		}
	} else {
		fmt.Println("Migrations applied successfully!")
	}
}
