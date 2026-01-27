//go:build scripts
// +build scripts

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		command    = flag.String("command", "up", "Migration command: up, down, force, version, or goto")
		version    = flag.Int("version", 0, "Version for force or goto commands")
		steps      = flag.Int("steps", 0, "Number of steps for up/down (0 = all)")
		dbHost     = flag.String("host", getEnv("DB_HOST", "localhost"), "Database host")
		dbPort     = flag.String("port", getEnv("DB_PORT", "5432"), "Database port")
		dbUser     = flag.String("user", getEnv("DB_USER", "postgres"), "Database user")
		dbPassword = flag.String("password", getEnv("DB_PASSWORD", "postgres"), "Database password")
		dbName     = flag.String("dbname", getEnv("DB_NAME", "equitywala"), "Database name")
		sslMode    = flag.String("sslmode", getEnv("DB_SSLMODE", "disable"), "SSL mode")
	)
	flag.Parse()

	// Build database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		*dbHost, *dbPort, *dbUser, *dbPassword, *dbName, *sslMode)

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Error pinging database: %v\n", err)
		os.Exit(1)
	}

	// Get migration directory path
	// Try multiple possible locations
	var migrationPath string
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Try paths relative to current working directory
	possiblePaths := []string{
		filepath.Join(cwd, "database", "migrations", "postgres"),           // From project root
		filepath.Join(cwd, "ssit-backend", "database", "migrations", "postgres"), // From workspace root
		filepath.Join(cwd, "..", "database", "migrations", "postgres"),  // From scripts directory
	}

	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			migrationPath = absPath
			break
		}
	}

	if migrationPath == "" {
		fmt.Fprintf(os.Stderr, "Migration directory not found. Tried:\n")
		for _, path := range possiblePaths {
			absPath, _ := filepath.Abs(path)
			fmt.Fprintf(os.Stderr, "  - %s\n", absPath)
		}
		os.Exit(1)
	}

	// Check if migration directory exists
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Migration directory not found: %s\n", migrationPath)
		os.Exit(1)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating postgres driver: %v\n", err)
		os.Exit(1)
	}

	// Create migrate instance
	// Use file:// protocol for local file system
	// Convert Windows backslashes to forward slashes for URL
	migrationURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationPath))
	m, err := migrate.NewWithDatabaseInstance(migrationURL, "postgres", driver)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating migrate instance: %v\n", err)
		os.Exit(1)
	}

	// Execute command
	switch *command {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "Error running migrations up: %v\n", err)
			os.Exit(1)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to apply")
		} else {
			fmt.Println("Migrations applied successfully")
		}

	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "Error running migrations down: %v\n", err)
			os.Exit(1)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to rollback")
		} else {
			fmt.Println("Migrations rolled back successfully")
		}

	case "force":
		if *version == 0 {
			fmt.Fprintf(os.Stderr, "Version is required for force command\n")
			os.Exit(1)
		}
		err = m.Force(*version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error forcing migration version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Migration version forced to %d\n", *version)

	case "goto":
		if *version == 0 {
			fmt.Fprintf(os.Stderr, "Version is required for goto command\n")
			os.Exit(1)
		}
		err = m.Migrate(uint(*version))
		if err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "Error migrating to version: %v\n", err)
			os.Exit(1)
		}
		if err == migrate.ErrNoChange {
			fmt.Printf("Already at version %d\n", *version)
		} else {
			fmt.Printf("Migrated to version %d\n", *version)
		}

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting migration version: %v\n", err)
			os.Exit(1)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		fmt.Fprintf(os.Stderr, "Available commands: up, down, force, goto, version\n")
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

