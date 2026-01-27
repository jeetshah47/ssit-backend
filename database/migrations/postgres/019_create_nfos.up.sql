-- Migration: Create NFOs Table

-- Ensure UUID extension is enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS nfos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    amc VARCHAR(255),
    category VARCHAR(100),
    type VARCHAR(50),
    open_date TIMESTAMP,
    close_date TIMESTAMP,
    timeline TEXT,
    summary TEXT, -- JSON array of summary points
    minimum_investment DECIMAL(10,2),
    verdict VARCHAR(50), -- Subscribe, Wait, Avoid
    rationale TEXT,
    report_url TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'upcoming', -- upcoming, open, closed
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_nfos_status ON nfos(status);
CREATE INDEX IF NOT EXISTS idx_nfos_open_date ON nfos(open_date);
CREATE INDEX IF NOT EXISTS idx_nfos_close_date ON nfos(close_date);
CREATE INDEX IF NOT EXISTS idx_nfos_created_at ON nfos(created_at DESC);

-- Add comments
COMMENT ON TABLE nfos IS 'New Fund Offers (NFOs) with advisory recommendations';
COMMENT ON COLUMN nfos.status IS 'Status: upcoming, open, closed';
COMMENT ON COLUMN nfos.verdict IS 'Advisory verdict: Subscribe, Wait, Avoid';
COMMENT ON COLUMN nfos.summary IS 'JSON array of summary points';
