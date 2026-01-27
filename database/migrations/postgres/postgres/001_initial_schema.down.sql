-- Migration: Rollback Initial Schema
-- Description: Drops all tables created in initial schema

-- Drop triggers first
DROP TRIGGER IF EXISTS update_team_members_updated_at ON team_members;
DROP TRIGGER IF EXISTS update_notification_settings_updated_at ON notification_settings;
DROP TRIGGER IF EXISTS update_advisories_updated_at ON advisories;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS update_vouchers_updated_at ON vouchers;
DROP TRIGGER IF EXISTS update_pricing_packages_updated_at ON pricing_packages;
DROP TRIGGER IF EXISTS update_kyc_records_updated_at ON kyc_records;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS user_classification_logs;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS notification_settings;
DROP TABLE IF EXISTS advisory_views;
DROP TABLE IF EXISTS advisories;
DROP TABLE IF EXISTS payment_webhooks;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS voucher_redemptions;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS vouchers;
DROP TABLE IF EXISTS pricing_packages;
DROP TABLE IF EXISTS kyc_records;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS otps;
DROP TABLE IF EXISTS users;

