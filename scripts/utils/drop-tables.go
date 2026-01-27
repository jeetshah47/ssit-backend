//go:build scripts
// +build scripts

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	var (
		confirm   = flag.Bool("confirm", false, "Confirm deletion (required for safety)")
		dbHost    = flag.String("host", getEnv("DB_HOST", "localhost"), "Database host")
		dbPort    = flag.String("port", getEnv("DB_PORT", "5432"), "Database port")
		dbUser    = flag.String("user", getEnv("DB_USER", "postgres"), "Database user")
		dbPassword = flag.String("password", getEnv("DB_PASSWORD", "postgres"), "Database password")
		dbName    = flag.String("dbname", getEnv("DB_NAME", "equitywala"), "Database name")
		sslMode   = flag.String("sslmode", getEnv("DB_SSLMODE", "disable"), "SSL mode")
	)
	flag.Parse()

	// Safety check
	if !*confirm {
		fmt.Fprintf(os.Stderr, "ERROR: This will DELETE ALL TABLES from the database!\n")
		fmt.Fprintf(os.Stderr, "Use -confirm flag to proceed.\n")
		os.Exit(1)
	}

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

	fmt.Printf("Connected to database: %s@%s:%s/%s\n", *dbUser, *dbHost, *dbPort, *dbName)
	fmt.Printf("WARNING: This will DROP ALL TABLES, TRIGGERS, and FUNCTIONS!\n")
	fmt.Printf("Proceeding...\n")

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting transaction: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback()

	// Drop triggers first (in reverse order of creation)
	// Each trigger is specific to its table
	triggerTablePairs := []struct {
		trigger string
		table   string
	}{
		{"update_team_members_updated_at", "team_members"},
		{"update_notification_settings_updated_at", "notification_settings"},
		{"update_advisories_updated_at", "advisories"},
		{"update_payments_updated_at", "payments"},
		{"update_subscriptions_updated_at", "subscriptions"},
		{"update_vouchers_updated_at", "vouchers"},
		{"update_pricing_packages_updated_at", "pricing_packages"},
		{"update_kyc_records_updated_at", "kyc_records"},
		{"update_roles_updated_at", "roles"},
		{"update_users_updated_at", "users"},
	}

	fmt.Println("\nDropping triggers...")
	for _, pair := range triggerTablePairs {
		query := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s CASCADE", pair.trigger, pair.table)
		if _, err := tx.Exec(query); err != nil {
			// Ignore errors - trigger might not exist
		} else {
			fmt.Printf("  Dropped trigger: %s on %s\n", pair.trigger, pair.table)
		}
	}

	// Drop function
	fmt.Println("Dropping functions...")
	if _, err := tx.Exec("DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error dropping function: %v\n", err)
	}

	// Drop tables in reverse order of dependencies
	tables := []string{
		"user_payment_plan_selections", // From migration 005
		"user_classification_logs",
		"team_members",
		"notification_settings",
		"advisory_views",
		"advisories",
		"payment_webhooks",
		"payments",
		"voucher_redemptions",
		"subscriptions",
		"vouchers",
		"pricing_packages",
		"kyc_records",
		"user_roles",
		"roles",
		"sessions",
		"otps",
		"users",
		"schema_migrations", // Migration tracking table
	}

	fmt.Println("\nDropping tables...")
	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := tx.Exec(query); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error dropping table %s: %v\n", table, err)
		} else {
			fmt.Printf("  Dropped table: %s\n", table)
		}
	}

	// Drop migration tracking table if it exists
	fmt.Println("\nDropping migration tracking...")
	if _, err := tx.Exec("DROP TABLE IF EXISTS schema_migrations CASCADE"); err != nil {
		// Ignore error - table might not exist
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "Error committing transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ All tables, triggers, and functions have been dropped successfully!")
	fmt.Println("Note: Migration version tracking has been reset. Run migrations to recreate tables.")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

