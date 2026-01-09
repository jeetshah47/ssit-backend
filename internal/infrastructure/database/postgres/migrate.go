package postgres

import (
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate runs GORM AutoMigrate for all models
func AutoMigrate(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create UUID extension: %w", err)
	}

	// Fix any existing tables with old UUID default syntax
	// This handles the case where tables were created with default:uuid_generate_v4() in GORM tags
	if err := fixExistingTables(db); err != nil {
		return fmt.Errorf("failed to fix existing tables: %w", err)
	}

	// Migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.Session{},
		&models.PaymentPlanSelection{},
	); err != nil {
		return fmt.Errorf("failed to migrate models: %w", err)
	}

	// TODO: Add other modules as they are implemented
	// - KYC models (kyc_records)
	// - Subscription models (pricing_packages, vouchers, subscriptions, voucher_redemptions)
	// - Payment models (payments, payment_webhooks)
	// - Advisory models (advisories, advisory_views)
	// - Settings models (notification_settings, team_members)
	// - Admin models (user_classification_logs)
	// - Roles models (roles, user_roles)

	return nil
}

// fixExistingTables fixes any existing tables that might have problematic default values
func fixExistingTables(db *gorm.DB) error {
	// Check if users table exists and fix ID column default if needed
	var exists bool
	if err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')").Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check if users table exists: %w", err)
	}

	if exists {
		// Remove any problematic default from ID column (UUIDs are handled by BeforeCreate hooks now)
		// This fixes the "insufficient arguments" error from old schema
		if err := db.Exec("ALTER TABLE users ALTER COLUMN id DROP DEFAULT IF EXISTS").Error; err != nil {
			// Ignore error if column doesn't have default or doesn't exist
		}
		if err := db.Exec("ALTER TABLE otps ALTER COLUMN id DROP DEFAULT IF EXISTS").Error; err != nil {
			// Ignore error if table doesn't exist yet
		}
		if err := db.Exec("ALTER TABLE sessions ALTER COLUMN id DROP DEFAULT IF EXISTS").Error; err != nil {
			// Ignore error if table doesn't exist yet
		}
	}

	return nil
}
