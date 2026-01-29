-- Migration: Rollback slow-query indexes
-- Description: Drops indexes added in 024_add_slow_query_indexes.up.sql

DROP INDEX IF EXISTS idx_stock_baskets_status_bullet_published;
DROP INDEX IF EXISTS idx_ipo_advisories_created_at;
