-- Migration: Create Sector Snapshots Table
-- Description: Creates sector_snapshots table for sector/index snapshots (NIFTY, BANK NIFTY, SENSEX, etc.)
-- Created: 2025-01-XX

CREATE TABLE sector_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sector_name VARCHAR(255) NOT NULL UNIQUE,
    current_value DECIMAL(10, 2),
    change_percentage DECIMAL(5, 2),
    change_value DECIMAL(10, 2),
    report_url TEXT,
    last_updated TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sector_snapshots_sector_name ON sector_snapshots(sector_name);
CREATE INDEX idx_sector_snapshots_last_updated ON sector_snapshots(last_updated DESC);

-- Insert default sectors
INSERT INTO sector_snapshots (sector_name) VALUES
    ('NIFTY'),
    ('BANK NIFTY'),
    ('SENSEX');
