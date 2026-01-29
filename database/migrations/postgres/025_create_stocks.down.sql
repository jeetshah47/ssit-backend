-- Migration: Drop stocks table
DROP INDEX IF EXISTS idx_stocks_exchange;
DROP INDEX IF EXISTS idx_stocks_symbol;
DROP INDEX IF EXISTS idx_stocks_name;
DROP TABLE IF EXISTS stocks;
