-- Migration: Add stock_id to stock_basket_items and remove redundant stock_name, stock_symbol
-- Description: Reference master stocks table instead of duplicating name/symbol on each item

-- 1. Add nullable FK to stocks
ALTER TABLE stock_basket_items
    ADD COLUMN stock_id UUID REFERENCES stocks(id);

-- 2. Ensure every distinct (stock_name, stock_symbol) exists in stocks
INSERT INTO stocks (id, name, symbol, exchange, created_at, updated_at)
SELECT uuid_generate_v4(), d.stock_name, d.stock_symbol, d.stock_symbol, NOW(), NOW()
FROM (
    SELECT DISTINCT stock_name, stock_symbol FROM stock_basket_items
) d
WHERE NOT EXISTS (
    SELECT 1 FROM stocks s
    WHERE s.name = d.stock_name AND (s.symbol IS NOT DISTINCT FROM d.stock_symbol)
);

-- 3. Backfill stock_id from stocks by name/symbol match
UPDATE stock_basket_items i
SET stock_id = (
    SELECT s.id FROM stocks s
    WHERE s.name = i.stock_name AND (s.symbol IS NOT DISTINCT FROM i.stock_symbol)
    LIMIT 1
);

-- 4. Make stock_id required
ALTER TABLE stock_basket_items
    ALTER COLUMN stock_id SET NOT NULL;

-- 5. Drop index on removed column
DROP INDEX IF EXISTS idx_stock_basket_items_stock_symbol;

-- 6. Remove redundant columns
ALTER TABLE stock_basket_items
    DROP COLUMN stock_name,
    DROP COLUMN stock_symbol;

-- 7. Index for lookups by stock
CREATE INDEX idx_stock_basket_items_stock_id ON stock_basket_items(stock_id);
