-- Migration: Create Weekly Market Mood Table

-- Ensure UUID extension is enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS weekly_market_mood (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    week VARCHAR(100) NOT NULL, -- e.g., "Week of Oct 21-27, 2024"
    points TEXT, -- JSON array of points
    sentiment DECIMAL(3,2) NOT NULL DEFAULT 0.5, -- 0.0 to 1.0
    sentiment_label VARCHAR(50) NOT NULL DEFAULT 'Neutral', -- Bullish, Bearish, Neutral
    report_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_weekly_market_mood_week ON weekly_market_mood(week);
CREATE INDEX IF NOT EXISTS idx_weekly_market_mood_created_at ON weekly_market_mood(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_weekly_market_mood_sentiment ON weekly_market_mood(sentiment);

-- Add comments
COMMENT ON TABLE weekly_market_mood IS 'Weekly market mood and sentiment analysis';
COMMENT ON COLUMN weekly_market_mood.sentiment IS 'Sentiment score from 0.0 (bearish) to 1.0 (bullish)';
COMMENT ON COLUMN weekly_market_mood.sentiment_label IS 'Human-readable sentiment: Bullish, Bearish, Neutral';
COMMENT ON COLUMN weekly_market_mood.points IS 'JSON array of key market points';
COMMENT ON COLUMN weekly_market_mood.week IS 'Week identifier (e.g., "Week of Oct 21-27, 2024")';
