package main

import (
	"database/sql"
	"log"

	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/server"
)

func main() {
	config.Must()
	db := SetupPostgres()

	server.New(db).Start()
}

func SetupPostgres() *sql.DB {
	db, err := sql.Open("postgres", config.Environment.DATABASE_URL)
	if err != nil {
		log.Fatalf("Unable to parse database url: %v", err)
	}

	return db
}
