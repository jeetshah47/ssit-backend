-- Migration: Add Payment Plan Fields
-- Description: Adds selected_payment_plan and selected_billing_period fields to users table for signup flow
-- Created: 2025-01-XX

-- Add payment plan fields
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS selected_payment_plan VARCHAR(50),
    ADD COLUMN IF NOT EXISTS selected_billing_period VARCHAR(50);

-- Add indexes for filtering
DROP INDEX IF EXISTS idx_users_selected_payment_plan;
CREATE INDEX idx_users_selected_payment_plan ON users(selected_payment_plan) WHERE selected_payment_plan IS NOT NULL;

DROP INDEX IF EXISTS idx_users_selected_billing_period;
CREATE INDEX idx_users_selected_billing_period ON users(selected_billing_period) WHERE selected_billing_period IS NOT NULL;

