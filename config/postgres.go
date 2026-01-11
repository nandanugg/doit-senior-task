package config

import (
	"database/sql"
	"fmt"

	// Import postgres driver for database/sql
	_ "github.com/lib/pq"
)

// NewPostgresDB creates and returns a new PostgreSQL database connection.
// It pings the database to ensure the connection is valid.
func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)

	return db, nil
}
