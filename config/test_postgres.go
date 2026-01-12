package config

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Import file source driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	// Import postgres driver for database/sql
	_ "github.com/lib/pq"
)

func init() {
	// Load .env file from project root
	envPath := filepath.Join(getProjectRoot(), ".env")
	_ = godotenv.Load(envPath)
}

// TestDB holds the test database connection and cleanup function.
type TestDB struct {
	DB      *sql.DB
	Cleanup func()
}

func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Get base connection string from environment
	baseConnStr := os.Getenv("TEST_DATABASE_URL")
	if baseConnStr == "" {
		t.Fatal("TEST_DATABASE_URL environment variable is required for integration tests")
	}

	// Connect to admin database to create test database
	adminDB, err := sql.Open("postgres", baseConnStr)
	if err != nil {
		t.Fatalf("failed to connect to admin database: %v", err)
	}

	// Test the connection
	if err := adminDB.Ping(); err != nil {
		_ = adminDB.Close()
		t.Fatalf("failed to ping admin database: %v", err)
	}

	// Generate unique database name using test name
	dbName := fmt.Sprintf("test_%s_%d", sanitizeDBName(t.Name()), os.Getpid())

	// Create test database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		_ = adminDB.Close()
		t.Fatalf("failed to create test database %s: %v", dbName, err)
	}

	// Build connection string for test database
	testConnStr, err := replaceDBName(baseConnStr, dbName)
	if err != nil {
		dropDatabase(adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("failed to build test database connection string: %v", err)
	}

	// Connect to test database
	testDB, err := sql.Open("postgres", testConnStr)
	if err != nil {
		dropDatabase(adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Ping to ensure connection is established
	if err := testDB.Ping(); err != nil {
		_ = testDB.Close()
		dropDatabase(adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Run migrations
	if err := runMigrations(testDB, dbName); err != nil {
		_ = testDB.Close()
		dropDatabase(adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Return TestDB with cleanup function
	cleanup := func() {
		_ = testDB.Close()
		dropDatabase(adminDB, dbName)
		_ = adminDB.Close()
	}

	return &TestDB{
		DB:      testDB,
		Cleanup: cleanup,
	}
}

// runMigrations runs all up migrations on the test database.
func runMigrations(db *sql.DB, dbName string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get the migrations directory path
	migrationsPath := getMigrationsPath()

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		dbName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// getProjectRoot returns the absolute path to the project root.
func getProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	// Navigate from config to project root
	return filepath.Join(filepath.Dir(filename), "..")
}

// getMigrationsPath returns the absolute path to the migrations directory.
func getMigrationsPath() string {
	return filepath.Join(getProjectRoot(), "migrations")
}

// dropDatabase drops the test database.
func dropDatabase(adminDB *sql.DB, dbName string) {
	// Terminate all connections to the database
	_, _ = adminDB.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		AND pid <> pg_backend_pid()
	`, dbName))

	// Drop the database
	_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
}

// sanitizeDBName converts a test name to a valid database name.
func sanitizeDBName(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if isAlphanumeric(c) {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	// Ensure it starts with a letter and is not too long
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = append([]byte{'t'}, result...)
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return strings.ToLower(string(result))
}

// isAlphanumeric checks if a character is alphanumeric.
func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// replaceDBName replaces the database name in a connection string.
func replaceDBName(connStr, newDBName string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Replace the path (database name)
	u.Path = "/" + newDBName

	return u.String(), nil
}
