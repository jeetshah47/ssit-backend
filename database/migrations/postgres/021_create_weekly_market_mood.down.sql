-- Migration: Rollback Create Weekly Market Mood Table

DROP INDEX IF EXISTS idx_weekly_market_mood_sentiment;
DROP INDEX IF EXISTS idx_weekly_market_mood_created_at;
DROP INDEX IF EXISTS idx_weekly_market_mood_week;

DROP TABLE IF EXISTS weekly_market_mood;
