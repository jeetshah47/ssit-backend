-- Migration: Add Role to Users (Rollback)
-- Description: Removes role column from users table

DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS role;

