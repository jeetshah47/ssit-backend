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
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// PricingPackage represents a pricing package to be inserted
type PricingPackage struct {
	Name         string
	Description  string
	Price        float64
	Currency     string
	DurationDays int
	DurationType string
	AccessLevel  string
	Status       string
	IsPublished  bool
}

func main() {
	// Command line flags
	var (
		envFile = flag.String("env", ".env", "Path to .env file")
		force   = flag.Bool("force", false, "Delete existing packages before inserting")
		clear   = flag.Bool("clear", false, "Clear all existing packages before inserting")
	)
	flag.Parse()

	// Load environment variables
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Get database connection string
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

	fmt.Println("Connected to database successfully")

	// Clear existing packages if requested
	if *clear {
		fmt.Println("Clearing all existing pricing packages...")
		_, err := db.Exec("DELETE FROM pricing_packages")
		if err != nil {
			log.Fatalf("Failed to clear pricing packages: %v", err)
		}
		fmt.Println("All pricing packages cleared")
	}

	// Define pricing packages
	packages := []PricingPackage{
		// Standard Plan - Quarterly
		{
			Name:         "Standard",
			Description:  "Basic advisory package with monthly reports and email support",
			Price:        249.75, // ~₹999/year = ₹249.75/quarter
			Currency:     "INR",
			DurationDays: 90,
			DurationType: "quarterly",
			AccessLevel:  "standard",
			Status:       "active",
			IsPublished:  true,
		},
		// Standard Plan - Annual
		{
			Name:         "Standard",
			Description:  "Basic advisory package with monthly reports and email support",
			Price:        999.00,
			Currency:     "INR",
			DurationDays: 365,
			DurationType: "annual",
			AccessLevel:  "standard",
			Status:       "active",
			IsPublished:  true,
		},
		// Plus Plan - Quarterly
		{
			Name:         "Plus",
			Description:  "Premium advisory with weekly reports, priority support, and stock baskets",
			Price:        499.75, // ~₹1999/year = ₹499.75/quarter
			Currency:     "INR",
			DurationDays: 90,
			DurationType: "quarterly",
			AccessLevel:  "plus",
			Status:       "active",
			IsPublished:  true,
		},
		// Plus Plan - Annual (Recommended)
		{
			Name:         "Plus",
			Description:  "Premium advisory with weekly reports, priority support, and stock baskets",
			Price:        1999.00,
			Currency:     "INR",
			DurationDays: 365,
			DurationType: "annual",
			AccessLevel:  "plus",
			Status:       "active",
			IsPublished:  true,
		},
		// Premium Plan - Quarterly
		{
			Name:         "Premium",
			Description:  "All Plus features plus daily reports, 24/7 support, personal advisor, and IPO access",
			Price:        749.75, // ~₹2999/year = ₹749.75/quarter
			Currency:     "INR",
			DurationDays: 90,
			DurationType: "quarterly",
			AccessLevel:  "premium",
			Status:       "active",
			IsPublished:  true,
		},
		// Premium Plan - Annual
		{
			Name:         "Premium",
			Description:  "All Plus features plus daily reports, 24/7 support, personal advisor, and IPO access",
			Price:        2999.00,
			Currency:     "INR",
			DurationDays: 365,
			DurationType: "annual",
			AccessLevel:  "premium",
			Status:       "active",
			IsPublished:  true,
		},
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	inserted := 0
	updated := 0
	skipped := 0

	for _, pkg := range packages {
		// Check if package already exists
		var existingID uuid.UUID
		err := tx.QueryRow(
			"SELECT id FROM pricing_packages WHERE LOWER(name) = LOWER($1) AND duration_type = $2",
			pkg.Name,
			pkg.DurationType,
		).Scan(&existingID)

		if err == nil {
			// Package exists
			if *force {
				// Update existing package
				_, err = tx.Exec(
					`UPDATE pricing_packages 
					SET description = $1, price = $2, currency = $3, duration_days = $4, 
					    access_level = $5, status = $6, is_published = $7, updated_at = $8
					WHERE id = $9`,
					pkg.Description,
					pkg.Price,
					pkg.Currency,
					pkg.DurationDays,
					pkg.AccessLevel,
					pkg.Status,
					pkg.IsPublished,
					time.Now(),
					existingID,
				)
				if err != nil {
					log.Printf("Failed to update package %s (%s): %v", pkg.Name, pkg.DurationType, err)
					continue
				}
				fmt.Printf("✓ Updated: %s (%s) - ₹%.2f\n", pkg.Name, pkg.DurationType, pkg.Price)
				updated++
			} else {
				fmt.Printf("⊘ Skipped: %s (%s) - already exists (use -force to update)\n", pkg.Name, pkg.DurationType)
				skipped++
			}
		} else if err == sql.ErrNoRows {
			// Package doesn't exist, insert it
			pkgID := uuid.New()
			now := time.Now()
			_, err = tx.Exec(
				`INSERT INTO pricing_packages 
				(id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
				pkgID,
				pkg.Name,
				pkg.Description,
				pkg.Price,
				pkg.Currency,
				pkg.DurationDays,
				pkg.DurationType,
				pkg.AccessLevel,
				pkg.Status,
				pkg.IsPublished,
				now,
				now,
			)
			if err != nil {
				log.Printf("Failed to insert package %s (%s): %v", pkg.Name, pkg.DurationType, err)
				continue
			}
			fmt.Printf("✓ Inserted: %s (%s) - ₹%.2f\n", pkg.Name, pkg.DurationType, pkg.Price)
			inserted++
		} else {
			log.Printf("Failed to check package existence %s (%s): %v", pkg.Name, pkg.DurationType, err)
			continue
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("Summary:\n")
	fmt.Printf("  Inserted: %d\n", inserted)
	fmt.Printf("  Updated:  %d\n", updated)
	fmt.Printf("  Skipped:  %d\n", skipped)
	fmt.Printf("  Total:    %d packages\n", len(packages))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Pricing packages setup completed successfully!")
}

// getDSN constructs database connection string from environment variables
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

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
