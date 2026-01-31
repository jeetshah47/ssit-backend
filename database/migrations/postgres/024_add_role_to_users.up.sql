-- Migration: Add Role to Users
-- Description: Adds role column to users table
-- Created: 2026-01-30

ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'user';
CREATE INDEX idx_users_role ON users(role);

