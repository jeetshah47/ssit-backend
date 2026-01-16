-- Migration: Drop Mutual Fund Baskets Tables
-- Description: Removes mutual_fund_baskets and mutual_fund_basket_items tables

DROP TABLE IF EXISTS mutual_fund_basket_items;
DROP TABLE IF EXISTS mutual_fund_baskets;
