-- Migration: Remove Payment Plan Fields from Users Table
-- Description: Removes denormalized selected_payment_plan and selected_billing_period fields from users table
-- Created: 2025-01-XX

-- Drop indexes
DROP INDEX IF EXISTS idx_users_selected_billing_period;
DROP INDEX IF EXISTS idx_users_selected_payment_plan;

-- Remove denormalized payment plan fields
ALTER TABLE users 
    DROP COLUMN IF EXISTS selected_payment_plan,
    DROP COLUMN IF EXISTS selected_billing_period;

