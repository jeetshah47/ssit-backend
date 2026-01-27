-- Migration: Remove Discount and Voucher Fields
-- Description: Removes discount and voucher-related fields from subscriptions table
--              and drops voucher-related tables
--              This migration is idempotent and handles both fresh installs and upgrades
--              Note: Initial schema (001) already has the correct structure with 'price' column
--              This migration only removes old columns if they exist (for upgrade scenarios)

-- Step 1: Drop voucher_redemptions table (depends on vouchers and subscriptions)
DROP TABLE IF EXISTS voucher_redemptions CASCADE;

-- Step 2: Drop vouchers table trigger
DROP TRIGGER IF EXISTS update_vouchers_updated_at ON vouchers;

-- Step 3: Drop vouchers table
DROP TABLE IF EXISTS vouchers CASCADE;

-- Step 4: Alter subscriptions table
-- Remove discount_amount, voucher_id, voucher_code columns if they exist
-- Change original_price and final_price to single price column (if they exist)
-- Update access_type constraint to remove 'free_voucher'

-- First, drop the existing constraints (will be recreated if needed)
ALTER TABLE subscriptions 
    DROP CONSTRAINT IF EXISTS subscriptions_access_type_check;

ALTER TABLE subscriptions 
    DROP CONSTRAINT IF EXISTS subscriptions_price_check;

-- Migrate data and drop old columns if they exist (for upgrade scenarios)
DO $$
DECLARE
    has_final_price BOOLEAN;
    has_original_price BOOLEAN;
BEGIN
    -- Check if old columns exist
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'subscriptions' AND column_name = 'final_price'
    ) INTO has_final_price;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'subscriptions' AND column_name = 'original_price'
    ) INTO has_original_price;

    -- Only migrate if old columns exist (upgrade scenario)
    IF has_final_price OR has_original_price THEN
        -- Ensure price column exists
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' AND table_name = 'subscriptions' AND column_name = 'price'
        ) THEN
            ALTER TABLE subscriptions ADD COLUMN price DECIMAL(10, 2);
        END IF;

        -- Migrate data: copy final_price to price (or original_price if final_price is null)
        IF has_final_price THEN
            EXECUTE 'UPDATE subscriptions SET price = final_price WHERE price IS NULL';
        ELSIF has_original_price THEN
            EXECUTE 'UPDATE subscriptions SET price = original_price WHERE price IS NULL';
        END IF;

        -- Set default for any remaining NULL values
        UPDATE subscriptions SET price = 0 WHERE price IS NULL;
        
        -- Make price NOT NULL
        ALTER TABLE subscriptions ALTER COLUMN price SET NOT NULL;

        -- Drop old columns
        IF has_original_price THEN
            ALTER TABLE subscriptions DROP COLUMN original_price;
        END IF;
        IF has_final_price THEN
            ALTER TABLE subscriptions DROP COLUMN final_price;
        END IF;
    END IF;
    
    -- Drop other old columns if they exist
    ALTER TABLE subscriptions DROP COLUMN IF EXISTS discount_amount;
    ALTER TABLE subscriptions DROP COLUMN IF EXISTS voucher_id;
    ALTER TABLE subscriptions DROP COLUMN IF EXISTS voucher_code;
END $$;

-- Recreate constraints with updated values
ALTER TABLE subscriptions 
    ADD CONSTRAINT subscriptions_access_type_check 
    CHECK (access_type IN ('paid', 'free_mf', 'admin_granted'));

ALTER TABLE subscriptions 
    ADD CONSTRAINT subscriptions_price_check 
    CHECK (price >= 0);

