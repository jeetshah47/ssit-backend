-- Migration: Create Weekly Audio Table

-- Ensure UUID extension is enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS weekly_audio (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    duration VARCHAR(50) NOT NULL, -- e.g., "45:30"
    week VARCHAR(100) NOT NULL, -- e.g., "Week of Oct 21-27, 2024"
    audio_url TEXT,
    thumbnail_url TEXT,
    posted_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'published', -- published, draft, archived
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_weekly_audio_status ON weekly_audio(status);
CREATE INDEX IF NOT EXISTS idx_weekly_audio_week ON weekly_audio(week);
CREATE INDEX IF NOT EXISTS idx_weekly_audio_posted_at ON weekly_audio(posted_at DESC);
CREATE INDEX IF NOT EXISTS idx_weekly_audio_created_at ON weekly_audio(created_at DESC);

-- Add comments
COMMENT ON TABLE weekly_audio IS 'Weekly audio content and podcasts';
COMMENT ON COLUMN weekly_audio.status IS 'Status: published, draft, archived';
COMMENT ON COLUMN weekly_audio.duration IS 'Duration in MM:SS format (e.g., "45:30")';
COMMENT ON COLUMN weekly_audio.week IS 'Week identifier (e.g., "Week of Oct 21-27, 2024")';
