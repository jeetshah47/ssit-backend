-- Migration: Rollback Add Payment Plan Fields
-- Description: Removes selected_payment_plan and selected_billing_period fields from users table

-- Drop indexes
DROP INDEX IF EXISTS idx_users_selected_billing_period;
DROP INDEX IF EXISTS idx_users_selected_payment_plan;

-- Remove columns
ALTER TABLE users 
    DROP COLUMN IF EXISTS selected_billing_period,
    DROP COLUMN IF EXISTS selected_payment_plan;

