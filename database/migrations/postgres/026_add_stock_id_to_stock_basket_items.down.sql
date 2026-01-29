-- Migration: Revert stock_id on stock_basket_items (restore stock_name, stock_symbol)
-- Description: Restore denormalized columns for backward compatibility

-- 1. Add columns back (nullable first)
ALTER TABLE stock_basket_items
    ADD COLUMN stock_name VARCHAR(255),
    ADD COLUMN stock_symbol VARCHAR(50);

-- 2. Populate from stocks
UPDATE stock_basket_items i
SET
    stock_name = s.name,
    stock_symbol = s.symbol
FROM stocks s
WHERE i.stock_id = s.id;

-- 3. Make stock_name NOT NULL (assume all rows have stock_id set)
ALTER TABLE stock_basket_items
    ALTER COLUMN stock_name SET NOT NULL;

-- 4. Drop FK column and index
DROP INDEX IF EXISTS idx_stock_basket_items_stock_id;
ALTER TABLE stock_basket_items
    DROP COLUMN stock_id;

-- 5. Restore index on symbol
CREATE INDEX idx_stock_basket_items_stock_symbol ON stock_basket_items(stock_symbol) WHERE stock_symbol IS NOT NULL;
