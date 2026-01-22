-- Migration: Rollback Create Webinars Table

DROP INDEX IF EXISTS idx_webinars_created_at;
DROP INDEX IF EXISTS idx_webinars_date;
DROP INDEX IF EXISTS idx_webinars_status;

DROP TABLE IF EXISTS webinars;
