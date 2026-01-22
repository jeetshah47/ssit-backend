-- Migration: Create MF Schemes Table

-- Ensure UUID extension is enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS mf_schemes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    scheme_code VARCHAR(50),
    amc VARCHAR(255),
    category VARCHAR(100),
    type VARCHAR(50),
    current_nav DECIMAL(10,4),
    entry_price DECIMAL(10,4),
    exit_price DECIMAL(10,4),
    trend VARCHAR(50),
    verdict VARCHAR(50),
    rationale TEXT,
    report_url TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_mf_schemes_status ON mf_schemes(status);
CREATE INDEX IF NOT EXISTS idx_mf_schemes_category ON mf_schemes(category);
CREATE INDEX IF NOT EXISTS idx_mf_schemes_type ON mf_schemes(type);
CREATE INDEX IF NOT EXISTS idx_mf_schemes_created_at ON mf_schemes(created_at DESC);

-- Add comments
COMMENT ON TABLE mf_schemes IS 'Mutual Fund Schemes recommended by the advisory';
COMMENT ON COLUMN mf_schemes.status IS 'Status: active, inactive';
COMMENT ON COLUMN mf_schemes.verdict IS 'Advisory verdict: Buy, Hold, Sell';
COMMENT ON COLUMN mf_schemes.trend IS 'Price trend: positive, negative, neutral';
