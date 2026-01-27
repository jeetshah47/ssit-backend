-- Migration: Rollback Add Customer Type Field
-- Description: Removes customer_type field from users table

-- Drop index
DROP INDEX IF EXISTS idx_users_customer_type;

-- Remove column
ALTER TABLE users 
    DROP COLUMN IF EXISTS customer_type;

