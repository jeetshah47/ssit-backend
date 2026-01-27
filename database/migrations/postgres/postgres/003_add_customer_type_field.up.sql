-- Migration: Add Customer Type Field
-- Description: Adds customer_type field to users table for Individual/Non Individual classification
-- Created: 2025-01-XX

-- Add customer_type column
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS customer_type VARCHAR(50);

-- Add index for filtering
DROP INDEX IF EXISTS idx_users_customer_type;
CREATE INDEX idx_users_customer_type ON users(customer_type) WHERE customer_type IS NOT NULL;

