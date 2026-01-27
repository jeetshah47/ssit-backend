-- Migration: Create IPO Advisories Table
-- Description: Creates ipo_advisories table for IPO advisory recommendations
-- Created: 2025-01-XX

CREATE TABLE ipo_advisories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advisory_type_id UUID NOT NULL REFERENCES advisory_types(id),
    ipo_name VARCHAR(255) NOT NULL,
    ipo_symbol VARCHAR(50),
    gmp DECIMAL(10, 2),
    suggestion VARCHAR(255),
    lot_size INTEGER NOT NULL,
    price_band_min DECIMAL(10, 2) NOT NULL,
    price_band_max DECIMAL(10, 2) NOT NULL,
    issue_date VARCHAR(100),
    issue_size VARCHAR(100),
    ipo_timetable TEXT,
    report_url TEXT,
    report_file_name VARCHAR(255),
    report_uploaded_at TIMESTAMP,
    report_uploaded_by UUID REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'upcoming',
    published_by UUID REFERENCES users(id),
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT ipo_advisories_status_check CHECK (status IN ('upcoming', 'open', 'closed', 'listed')),
    CONSTRAINT ipo_advisories_price_band_check CHECK (price_band_max >= price_band_min),
    CONSTRAINT ipo_advisories_lot_size_check CHECK (lot_size > 0)
);

CREATE INDEX idx_ipo_advisories_advisory_type_id ON ipo_advisories(advisory_type_id);
CREATE INDEX idx_ipo_advisories_status ON ipo_advisories(status);
CREATE INDEX idx_ipo_advisories_published_at ON ipo_advisories(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_ipo_advisories_published_by ON ipo_advisories(published_by);
CREATE INDEX idx_ipo_advisories_ipo_symbol ON ipo_advisories(ipo_symbol) WHERE ipo_symbol IS NOT NULL;
