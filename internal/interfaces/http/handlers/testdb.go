package handlers

import (
	"os"
	"testing"

	"github.com/equitywala/backend/internal/domain/auth"
	"github.com/equitywala/backend/internal/domain/user"
	postgresAuth "github.com/equitywala/backend/internal/infrastructure/database/postgres/auth"
	postgresDB "github.com/equitywala/backend/internal/infrastructure/database/postgres"
	postgresUser "github.com/equitywala/backend/internal/infrastructure/database/postgres/user"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getTestDBDSN returns the test database DSN from environment or uses default
func getTestDBDSN() string {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		// Default test database connection string
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=equitywala_test sslmode=disable"
	}
	return dsn
}

// setupTestDB creates a PostgreSQL database connection for testing
// Uses the same PostgreSQL driver as production
func setupTestDB(t testing.TB) *gorm.DB {
	dsn := getTestDBDSN()
	
	// Use the same connection function as production
	db, err := postgresDB.NewConnection(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v. Make sure PostgreSQL is running and TEST_DB_DSN is set correctly.", err)
	}

	// Set logger to silent for tests
	db.Logger = logger.Default.LogMode(logger.Silent)

	// Run migrations using AutoMigrate (same as production)
	if err := postgresDB.AutoMigrate(db); err != nil {
		postgresDB.CloseDB(db)
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupTestDBFile creates a PostgreSQL database connection for testing with cleanup
// Note: PostgreSQL doesn't use files like SQLite, but we provide cleanup function
func setupTestDBFile(t testing.TB) (*gorm.DB, func()) {
	dsn := getTestDBDSN()
	
	db, err := postgresDB.NewConnection(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set logger to silent for tests
	db.Logger = logger.Default.LogMode(logger.Silent)

	// Run migrations
	if err := postgresDB.AutoMigrate(db); err != nil {
		postgresDB.CloseDB(db)
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		// Clean up test data (optional - you can drop tables or truncate)
		// For now, just close the connection
		if err := postgresDB.CloseDB(db); err != nil {
			t.Logf("Error closing test database: %v", err)
		}
	}

	return db, cleanup
}

// setupTestRepositories creates all test repositories
func setupTestRepositories(db *gorm.DB) (
	user.Repository,
	auth.OTPRepository,
	auth.SessionRepository,
) {
	userRepo := postgresUser.NewRepository(db)
	otpRepo := postgresAuth.NewOTPRepository(db)
	sessionRepo := postgresAuth.NewSessionRepository(db)

	return userRepo, otpRepo, sessionRepo
}

