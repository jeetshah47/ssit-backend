-- Script to truncate all tables in the database
-- Usage: psql -U postgres -d equitywala -f truncate-tables.sql
-- Or: psql $DATABASE_URL -f truncate-tables.sql

-- Disable foreign key checks temporarily by truncating in correct order
-- Tables with foreign keys are truncated first

BEGIN;

-- Truncate tables in reverse dependency order (child tables first)
-- This ensures foreign key constraints are respected

-- Payment module (depends on users, subscriptions)
TRUNCATE TABLE payment_webhooks CASCADE;
TRUNCATE TABLE payments CASCADE;

-- Subscription module (depends on users, pricing_packages, vouchers)
TRUNCATE TABLE voucher_redemptions CASCADE;
TRUNCATE TABLE subscriptions CASCADE;
TRUNCATE TABLE vouchers CASCADE;
TRUNCATE TABLE pricing_packages CASCADE;

-- Advisory module (depends on users, advisories)
TRUNCATE TABLE advisory_views CASCADE;
TRUNCATE TABLE advisories CASCADE;

-- Settings module (depends on users, roles)
TRUNCATE TABLE team_members CASCADE;
TRUNCATE TABLE notification_settings CASCADE;

-- Admin module (depends on users)
TRUNCATE TABLE user_classification_logs CASCADE;

-- KYC module (depends on users)
TRUNCATE TABLE kyc_records CASCADE;

-- Auth module (depends on users, roles)
TRUNCATE TABLE user_roles CASCADE;
TRUNCATE TABLE sessions CASCADE;
TRUNCATE TABLE otps CASCADE;

-- Core tables (no dependencies on other application tables)
TRUNCATE TABLE roles CASCADE;
TRUNCATE TABLE users CASCADE;

-- Reset sequences if any (PostgreSQL auto-increment sequences)
-- Note: UUID primary keys don't use sequences, but if you have any serial/bigserial columns, reset them here
-- Example: ALTER SEQUENCE IF EXISTS table_name_id_seq RESTART WITH 1;

COMMIT;

-- Verify tables are empty (optional - uncomment to check)
-- SELECT 
--     'users' as table_name, COUNT(*) as row_count FROM users
-- UNION ALL
-- SELECT 'otps', COUNT(*) FROM otps
-- UNION ALL
-- SELECT 'sessions', COUNT(*) FROM sessions
-- UNION ALL
-- SELECT 'roles', COUNT(*) FROM roles
-- UNION ALL
-- SELECT 'user_roles', COUNT(*) FROM user_roles
-- UNION ALL
-- SELECT 'kyc_records', COUNT(*) FROM kyc_records
-- UNION ALL
-- SELECT 'pricing_packages', COUNT(*) FROM pricing_packages
-- UNION ALL
-- SELECT 'vouchers', COUNT(*) FROM vouchers
-- UNION ALL
-- SELECT 'subscriptions', COUNT(*) FROM subscriptions
-- UNION ALL
-- SELECT 'voucher_redemptions', COUNT(*) FROM voucher_redemptions
-- UNION ALL
-- SELECT 'payments', COUNT(*) FROM payments
-- UNION ALL
-- SELECT 'payment_webhooks', COUNT(*) FROM payment_webhooks
-- UNION ALL
-- SELECT 'advisories', COUNT(*) FROM advisories
-- UNION ALL
-- SELECT 'advisory_views', COUNT(*) FROM advisory_views
-- UNION ALL
-- SELECT 'notification_settings', COUNT(*) FROM notification_settings
-- UNION ALL
-- SELECT 'team_members', COUNT(*) FROM team_members
-- UNION ALL
-- SELECT 'user_classification_logs', COUNT(*) FROM user_classification_logs;

