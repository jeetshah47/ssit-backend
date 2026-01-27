//go:build scripts
// +build scripts

package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "equitywala")
	sslMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)

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

	// Check current constraint
	var constraintDef string
	err = db.QueryRow(`
		SELECT pg_get_constraintdef(oid) 
		FROM pg_constraint 
		WHERE conrelid = 'payments'::regclass 
		AND conname = 'payments_method_check'
	`).Scan(&constraintDef)

	if err == sql.ErrNoRows {
		fmt.Println("Constraint not found, creating it...")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking constraint: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Current constraint: %s\n", constraintDef)
	}

	// Drop and recreate constraint
	fmt.Println("Dropping existing constraint...")
	_, err = db.Exec(`ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_method_check`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dropping constraint: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating new constraint with 'paytm'...")
	_, err = db.Exec(`
		ALTER TABLE payments 
		ADD CONSTRAINT payments_method_check 
		CHECK (payment_method IN ('upi', 'card', 'netbanking', 'razorpay', 'paytm'))
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating constraint: %v\n", err)
		os.Exit(1)
	}

	// Verify the constraint
	err = db.QueryRow(`
		SELECT pg_get_constraintdef(oid) 
		FROM pg_constraint 
		WHERE conrelid = 'payments'::regclass 
		AND conname = 'payments_method_check'
	`).Scan(&constraintDef)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying constraint: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Constraint updated successfully!\n")
	fmt.Printf("New constraint: %s\n", constraintDef)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

