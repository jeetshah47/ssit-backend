-- Migration: Rollback Create MF Schemes Table

DROP INDEX IF EXISTS idx_mf_schemes_created_at;
DROP INDEX IF EXISTS idx_mf_schemes_type;
DROP INDEX IF EXISTS idx_mf_schemes_category;
DROP INDEX IF EXISTS idx_mf_schemes_status;

DROP TABLE IF EXISTS mf_schemes;
