package db

import (
	"log"
	"os"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func RunMigrations() {
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Fatal("POSTGRES_URL not set in environment")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not start postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("migration failed to start: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run up migrations: %v", err)
	}

	log.Println("✅ Migrations applied successfully!")
}
