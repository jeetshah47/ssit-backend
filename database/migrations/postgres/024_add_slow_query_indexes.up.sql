-- Migration: Add indexes to speed up slow queries
-- Description: Adds created_at index on ipo_advisories and composite index on stock_baskets
-- for ORDER BY created_at DESC and (status, is_bullet_idea, published_at DESC) queries

-- Speed up: SELECT * FROM ipo_advisories ORDER BY created_at DESC LIMIT N
CREATE INDEX idx_ipo_advisories_created_at ON ipo_advisories(created_at DESC);

-- Speed up: SELECT * FROM stock_baskets WHERE status = 'published' AND is_bullet_idea = true ORDER BY published_at DESC LIMIT 1
CREATE INDEX idx_stock_baskets_status_bullet_published ON stock_baskets(status, is_bullet_idea, published_at DESC NULLS LAST)
WHERE status = 'published' AND is_bullet_idea = true;
