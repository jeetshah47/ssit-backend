-- Migration: Drop ETF Baskets Tables
-- Description: Removes etf_baskets and etf_basket_items tables

DROP TABLE IF EXISTS etf_basket_items;
DROP TABLE IF EXISTS etf_baskets;
