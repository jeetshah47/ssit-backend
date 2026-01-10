-- Migration: Rollback Remove Payment Plan Fields from Users Table
-- Description: Restores selected_payment_plan and selected_billing_period fields to users table

-- Add payment plan fields back
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS selected_payment_plan VARCHAR(50),
    ADD COLUMN IF NOT EXISTS selected_billing_period VARCHAR(50);

-- Restore indexes
CREATE INDEX IF NOT EXISTS idx_users_selected_payment_plan ON users(selected_payment_plan) WHERE selected_payment_plan IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_selected_billing_period ON users(selected_billing_period) WHERE selected_billing_period IS NOT NULL;

