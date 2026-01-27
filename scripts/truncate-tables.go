//go:build scripts
// +build scripts

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	// Command line flags
	var (
		confirm     = flag.Bool("confirm", false, "Confirm truncation (required for safety)")
		tables      = flag.String("tables", "", "Comma-separated list of specific tables to truncate (empty = all tables)")
		verify      = flag.Bool("verify", false, "Verify tables are empty after truncation")
		envFile     = flag.String("env", ".env", "Path to .env file")
		showTables  = flag.Bool("list", false, "List all tables that would be truncated")
	)
	flag.Parse()

	// Load environment variables
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Get database connection string from environment
	dsn := getDSN()
	if dsn == "" {
		log.Fatal("Database connection string not found. Set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME environment variables or use DATABASE_URL")
	}

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Define all tables in truncation order (child tables first)
	allTables := []string{
		"payment_webhooks",
		"payments",
		"voucher_redemptions",
		"subscriptions",
		"vouchers",
		"pricing_packages",
		"advisory_views",
		"advisories",
		"team_members",
		"notification_settings",
		"user_classification_logs",
		"kyc_records",
		"user_roles",
		"sessions",
		"otps",
		"roles",
		"users",
	}

	// Determine which tables to truncate
	tablesToTruncate := allTables
	if *tables != "" {
		requestedTables := strings.Split(*tables, ",")
		for i := range requestedTables {
			requestedTables[i] = strings.TrimSpace(requestedTables[i])
		}
		tablesToTruncate = requestedTables
	}

	// List tables if requested
	if *showTables {
		fmt.Println("Tables that would be truncated:")
		for _, table := range tablesToTruncate {
			count, _ := getTableCount(db, table)
			fmt.Printf("  - %s (%d rows)\n", table, count)
		}
		return
	}

	// Safety check
	if !*confirm {
		fmt.Println("ERROR: Truncation requires confirmation flag")
		fmt.Println("Usage: go run truncate-tables.go -confirm")
		fmt.Println("Or: go run truncate-tables.go -confirm -tables=users,otps")
		fmt.Println("Use -list to see tables and row counts")
		os.Exit(1)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Truncate tables
	fmt.Println("Truncating tables...")
	for _, table := range tablesToTruncate {
		// Check if table exists
		var exists bool
		err := tx.QueryRow(
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
			table,
		).Scan(&exists)
		if err != nil {
			log.Printf("Warning: Could not check if table %s exists: %v", table, err)
			continue
		}
		if !exists {
			log.Printf("Warning: Table %s does not exist, skipping", table)
			continue
		}

		// Get row count before truncation
		count, _ := getTableCountTx(tx, table)
		if count > 0 {
			fmt.Printf("  Truncating %s (%d rows)...\n", table, count)
		} else {
			fmt.Printf("  Truncating %s (already empty)...\n", table)
		}

		// Truncate table
		_, err = tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("\n✓ All tables truncated successfully!")

	// Verify if requested
	if *verify {
		fmt.Println("\nVerifying tables are empty...")
		allEmpty := true
		for _, table := range tablesToTruncate {
			count, err := getTableCount(db, table)
			if err != nil {
				log.Printf("Warning: Could not verify table %s: %v", table, err)
				continue
			}
			if count > 0 {
				fmt.Printf("  ✗ %s: %d rows (not empty!)\n", table, count)
				allEmpty = false
			} else {
				fmt.Printf("  ✓ %s: empty\n", table)
			}
		}
		if allEmpty {
			fmt.Println("\n✓ All tables verified as empty!")
		} else {
			fmt.Println("\n✗ Some tables are not empty!")
			os.Exit(1)
		}
	}
}

func getDSN() string {
	// Try DATABASE_URL first
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	// Build from individual components
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "equitywala")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if host == "" || user == "" || dbname == "" {
		return ""
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getTableCount(db *sql.DB, table string) (int, error) {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	return count, err
}

func getTableCountTx(tx *sql.Tx, table string) (int, error) {
	var count int
	err := tx.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	return count, err
}

