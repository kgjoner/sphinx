package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

// CleanDatabase truncates all tables for a clean test state
func CleanDatabase(dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Get all tables in the public schema
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'schema_migrations'
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %v", err)
	}

	// Truncate all tables
	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE \"%s\" CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %v", table, err)
		}
	}

	log.Println("Database cleaned successfully!")
	return nil
}
