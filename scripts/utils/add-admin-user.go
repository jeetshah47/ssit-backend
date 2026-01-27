//go:build scripts
// +build scripts

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

func main() {
	// Command line flags
	var (
		email    = flag.String("email", "", "Admin user email (required)")
		name     = flag.String("name", "", "Admin user name (required)")
		password = flag.String("password", "", "Admin user password (required)")
		roleName = flag.String("role", "admin", "Role name to assign (default: 'admin')")
		envFile  = flag.String("env", ".env", "Path to .env file")
		force    = flag.Bool("force", false, "Force creation even if user already exists")
	)
	flag.Parse()

	// Validate required flags
	if *email == "" || *name == "" || *password == "" {
		fmt.Println("ERROR: Missing required parameters")
		fmt.Println("Usage: go run add-admin-user.go -email=<email> -name=<name> -password=<password>")
		fmt.Println("Optional flags:")
		fmt.Println("  -role=<role>     Role name (default: 'admin')")
		fmt.Println("  -env=<path>       Path to .env file (default: '.env')")
		fmt.Println("  -force            Force creation even if user exists")
		os.Exit(1)
	}

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
	fmt.Printf("Creating admin user: %s (%s)\n", *name, *email)

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Check if user already exists
	var existingUserID string
	err = tx.QueryRow("SELECT id FROM users WHERE email = $1", *email).Scan(&existingUserID)
	if err == nil {
		if !*force {
			log.Fatalf("User with email %s already exists. Use -force to update existing user.", *email)
		}
		fmt.Printf("User already exists, updating...\n")
	} else if err != sql.ErrNoRows {
		log.Fatalf("Failed to check user existence: %v", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	now := time.Now()
	var userID uuid.UUID

	if existingUserID != "" {
		// Update existing user
		userID, err = uuid.Parse(existingUserID)
		if err != nil {
			log.Fatalf("Failed to parse user ID: %v", err)
		}

		_, err = tx.Exec(`
			UPDATE users 
			SET name = $1, 
			    password_hash = $2, 
			    status = $3, 
			    email_verified = $4, 
			    email_verified_at = $5,
			    updated_at = $6
			WHERE id = $7
		`, *name, string(passwordHash), "active", true, now, now, userID)
		if err != nil {
			log.Fatalf("Failed to update user: %v", err)
		}
		fmt.Println("✓ User updated successfully")
	} else {
		// Create new user
		userID = uuid.New()
		_, err = tx.Exec(`
			INSERT INTO users (id, email, name, password_hash, status, email_verified, email_verified_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, userID, *email, *name, string(passwordHash), "active", true, now, now, now)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		fmt.Println("✓ User created successfully")
	}

	// Ensure role exists
	var roleID uuid.UUID
	err = tx.QueryRow("SELECT id FROM roles WHERE name = $1", *roleName).Scan(&roleID)
	if err == sql.ErrNoRows {
		// Create role if it doesn't exist
		roleID = uuid.New()
		_, err = tx.Exec(`
			INSERT INTO roles (id, name, description, is_system_role, permissions, member_count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, roleID, *roleName, fmt.Sprintf("System role for %s users", *roleName), true, "[]", 0, now, now)
		if err != nil {
			log.Fatalf("Failed to create role: %v", err)
		}
		fmt.Printf("✓ Role '%s' created\n", *roleName)
	} else if err != nil {
		log.Fatalf("Failed to check role existence: %v", err)
	} else {
		fmt.Printf("✓ Role '%s' found\n", *roleName)
	}

	// Check if role is already assigned
	var existingUserRoleID string
	err = tx.QueryRow("SELECT id FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, roleID).Scan(&existingUserRoleID)
	if err == sql.ErrNoRows {
		// Assign role to user
		_, err = tx.Exec(`
			INSERT INTO user_roles (id, user_id, role_id, assigned_at)
			VALUES ($1, $2, $3, $4)
		`, uuid.New(), userID, roleID, now)
		if err != nil {
			log.Fatalf("Failed to assign role to user: %v", err)
		}
		fmt.Printf("✓ Role '%s' assigned to user\n", *roleName)
	} else if err != nil {
		log.Fatalf("Failed to check role assignment: %v", err)
	} else {
		fmt.Printf("✓ Role '%s' already assigned to user\n", *roleName)
	}

	// Update role member count
	_, err = tx.Exec(`
		UPDATE roles 
		SET member_count = (
			SELECT COUNT(*) FROM user_roles WHERE role_id = $1
		),
		updated_at = $2
		WHERE id = $1
	`, roleID, now)
	if err != nil {
		log.Printf("Warning: Failed to update role member count: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("\n✓ Admin user created/updated successfully!")
	fmt.Printf("  User ID: %s\n", userID.String())
	fmt.Printf("  Email: %s\n", *email)
	fmt.Printf("  Name: %s\n", *name)
	fmt.Printf("  Role: %s\n", *roleName)
	fmt.Printf("  Status: active\n")
	fmt.Printf("  Email Verified: true\n")
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
