-- Migration: Rollback Create Weekly Audio Table

DROP INDEX IF EXISTS idx_weekly_audio_created_at;
DROP INDEX IF EXISTS idx_weekly_audio_posted_at;
DROP INDEX IF EXISTS idx_weekly_audio_week;
DROP INDEX IF EXISTS idx_weekly_audio_status;

DROP TABLE IF EXISTS weekly_audio;
