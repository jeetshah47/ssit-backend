//go:build scripts
// +build scripts

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	var (
		envFile    = flag.String("env", ".env", "Path to .env file")
		command    = flag.String("command", "up", "Migration command: up, down, force, version, or goto")
		version    = flag.Int("version", 0, "Version for force or goto commands")
		steps      = flag.Int("steps", 0, "Number of steps for up/down (0 = all)")
		dbHost     = flag.String("host", "", "Database host (overrides .env)")
		dbPort     = flag.String("port", "", "Database port (overrides .env)")
		dbUser     = flag.String("user", "", "Database user (overrides .env)")
		dbPassword = flag.String("password", "", "Database password (overrides .env)")
		dbName     = flag.String("dbname", "", "Database name (overrides .env)")
		sslMode    = flag.String("sslmode", "", "SSL mode (overrides .env)")
	)
	flag.Parse()

	// Load environment variables from .env file
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Get database configuration (command line flags override .env)
	if *dbHost == "" {
		*dbHost = getEnv("DB_HOST", "localhost")
	}
	if *dbPort == "" {
		*dbPort = getEnv("DB_PORT", "5432")
	}
	if *dbUser == "" {
		*dbUser = getEnv("DB_USER", "postgres")
	}
	if *dbPassword == "" {
		*dbPassword = getEnv("DB_PASSWORD", "postgres")
	}
	if *dbName == "" {
		*dbName = getEnv("DB_NAME", "equitywala")
	}
	if *sslMode == "" {
		*sslMode = getEnv("DB_SSLMODE", "disable")
	}

	// Debug: Print connection parameters (without password)
	fmt.Printf("Database connection parameters:\n")
	fmt.Printf("  Host: %s\n", *dbHost)
	fmt.Printf("  Port: %s\n", *dbPort)
	fmt.Printf("  User: %s\n", *dbUser)
	fmt.Printf("  Database: %s\n", *dbName)
	fmt.Printf("  SSL Mode: %s\n", *sslMode)
	fmt.Println()

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
		filepath.Join(cwd, "database", "migrations", "postgres"),                 // From project root
		filepath.Join(cwd, "ssit-backend", "database", "migrations", "postgres"), // From workspace root
		filepath.Join(cwd, "..", "database", "migrations", "postgres"),           // From scripts directory
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
