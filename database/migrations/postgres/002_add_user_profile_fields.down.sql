-- Migration: Rollback Add User Profile Fields
-- Description: Removes profile fields and restores password_hash NOT NULL constraint

-- Drop indexes
DROP INDEX IF EXISTS idx_users_date_of_birth;
DROP INDEX IF EXISTS idx_users_city;
DROP INDEX IF EXISTS idx_users_pan_unique;

-- Remove new columns
ALTER TABLE users 
    DROP COLUMN IF EXISTS pan_mobile,
    DROP COLUMN IF EXISTS pan_address,
    DROP COLUMN IF EXISTS pan_name,
    DROP COLUMN IF EXISTS pan,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS date_of_birth;

-- Restore password_hash NOT NULL constraint
-- WARNING: This will fail if there are any NULL values in password_hash
-- You may need to update existing records first or handle this manually
-- Uncomment the following line only after ensuring all password_hash values are NOT NULL
-- ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

