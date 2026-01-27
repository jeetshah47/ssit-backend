-- Migration: Create Webinars Table

-- Ensure UUID extension is enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS webinars (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'upcoming', -- Live, Upcoming, Recorded
    date TIMESTAMP,
    time VARCHAR(50), -- HH:MM format
    duration VARCHAR(50), -- e.g., "60 minutes"
    registration_url TEXT,
    recording_url TEXT,
    thumbnail_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_webinars_status ON webinars(status);
CREATE INDEX IF NOT EXISTS idx_webinars_date ON webinars(date);
CREATE INDEX IF NOT EXISTS idx_webinars_created_at ON webinars(created_at DESC);

-- Add comments
COMMENT ON TABLE webinars IS 'Webinars and live sessions';
COMMENT ON COLUMN webinars.status IS 'Status: Live, Upcoming, Recorded';
COMMENT ON COLUMN webinars.time IS 'Time in HH:MM format (24-hour)';
COMMENT ON COLUMN webinars.duration IS 'Duration in human-readable format (e.g., "60 minutes", "1.5 hours")';
