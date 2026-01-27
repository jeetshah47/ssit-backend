-- Migration: Rollback Create User Payment Plan Selections Table
-- Description: Removes user_payment_plan_selections table

-- Drop indexes
DROP INDEX IF EXISTS idx_user_payment_plan_selections_status;
DROP INDEX IF EXISTS idx_user_payment_plan_selections_package_id;
DROP INDEX IF EXISTS idx_user_payment_plan_selections_user_id;

-- Drop table
DROP TABLE IF EXISTS user_payment_plan_selections;

