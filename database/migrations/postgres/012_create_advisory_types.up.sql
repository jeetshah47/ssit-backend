-- Migration: Create Advisory Types Table
-- Description: Creates master table for advisory types (Stock Basket, ETF Basket, IPO, Mutual Fund Basket, etc.)
-- Created: 2025-01-XX

CREATE TABLE advisory_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_advisory_types_name ON advisory_types(name);
CREATE INDEX idx_advisory_types_is_active ON advisory_types(is_active) WHERE is_active = true;

-- Insert default advisory types
INSERT INTO advisory_types (name, display_name, description) VALUES
    ('stock_basket', 'Stock Basket', 'Basket of stock recommendations'),
    ('etf_basket', 'ETF Basket', 'Basket of ETF recommendations'),
    ('ipo', 'IPO Advisory', 'Initial Public Offering advisory'),
    ('mutual_fund_basket', 'Mutual Fund Basket', 'Basket of mutual fund recommendations'),
    ('commodities', 'Commodities', 'Commodities advisory'),
    ('pms_aif', 'PMS/AIF', 'Portfolio Management Services / Alternative Investment Funds'),
    ('sif', 'SIF', 'Structured Investment Funds');
