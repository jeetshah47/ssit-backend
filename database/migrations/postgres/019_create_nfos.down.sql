-- Migration: Rollback Create NFOs Table

DROP INDEX IF EXISTS idx_nfos_created_at;
DROP INDEX IF EXISTS idx_nfos_close_date;
DROP INDEX IF EXISTS idx_nfos_open_date;
DROP INDEX IF EXISTS idx_nfos_status;

DROP TABLE IF EXISTS nfos;
