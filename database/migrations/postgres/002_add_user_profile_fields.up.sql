-- Migration: Add User Profile Fields
-- Description: Adds profile fields for step-by-step signup (DOB, City, PAN) and makes password_hash nullable
-- Created: 2025-01-XX

-- Make password_hash nullable (for step-by-step signup where password is set later)
ALTER TABLE users 
    ALTER COLUMN password_hash DROP NOT NULL;

-- Add new profile fields
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS date_of_birth DATE,
    ADD COLUMN IF NOT EXISTS city VARCHAR(100),
    ADD COLUMN IF NOT EXISTS pan VARCHAR(10),
    ADD COLUMN IF NOT EXISTS pan_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS pan_address TEXT,
    ADD COLUMN IF NOT EXISTS pan_mobile VARCHAR(20);

-- Add unique index on PAN (only for non-null values)
-- Drop index first if it exists to avoid conflicts
DROP INDEX IF EXISTS idx_users_pan_unique;
CREATE UNIQUE INDEX idx_users_pan_unique ON users(pan) WHERE pan IS NOT NULL;

-- Add index on city for filtering/searching
DROP INDEX IF EXISTS idx_users_city;
CREATE INDEX idx_users_city ON users(city) WHERE city IS NOT NULL;

-- Add index on date_of_birth for age-based queries
DROP INDEX IF EXISTS idx_users_date_of_birth;
CREATE INDEX idx_users_date_of_birth ON users(date_of_birth) WHERE date_of_birth IS NOT NULL;

