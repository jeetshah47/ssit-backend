-- Migration: Rollback Remove Discount and Voucher Fields
-- Description: Restores discount and voucher-related fields and tables

-- Step 1: Recreate vouchers table
CREATE TABLE IF NOT EXISTS vouchers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    
    -- Discount Type
    discount_type VARCHAR(50) NOT NULL,
    discount_value DECIMAL(10, 2) NOT NULL,
    
    -- Validity
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    
    -- Usage Limits
    max_uses INTEGER,
    used_count INTEGER NOT NULL DEFAULT 0,
    max_uses_per_user INTEGER DEFAULT 1,
    
    -- Applicability
    applicable_packages JSONB,
    applicable_user_types JSONB,
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Metadata
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT vouchers_discount_type_check CHECK (discount_type IN ('percentage', 'flat', 'free')),
    CONSTRAINT vouchers_discount_value_check CHECK (
        (discount_type = 'percentage' AND discount_value >= 0 AND discount_value <= 100) OR
        (discount_type = 'flat' AND discount_value >= 0) OR
        (discount_type = 'free' AND discount_value = 100)
    ),
    CONSTRAINT vouchers_validity_check CHECK (valid_until > valid_from)
);

CREATE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers(code);
CREATE INDEX IF NOT EXISTS idx_vouchers_active ON vouchers(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_vouchers_validity ON vouchers(valid_from, valid_until);

-- Step 2: Recreate voucher trigger
CREATE TRIGGER update_vouchers_updated_at BEFORE UPDATE ON vouchers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Step 3: Recreate voucher_redemptions table
CREATE TABLE IF NOT EXISTS voucher_redemptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    voucher_id UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    
    discount_amount DECIMAL(10, 2) NOT NULL,
    redeemed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT voucher_redemptions_unique UNIQUE (voucher_id, user_id, subscription_id)
);

CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_voucher_id ON voucher_redemptions(voucher_id);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_user_id ON voucher_redemptions(user_id);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_redeemed_at ON voucher_redemptions(redeemed_at);

-- Step 4: Alter subscriptions table to restore old structure
-- Drop existing constraints
ALTER TABLE subscriptions 
    DROP CONSTRAINT IF EXISTS subscriptions_access_type_check;

ALTER TABLE subscriptions 
    DROP CONSTRAINT IF EXISTS subscriptions_price_check;

-- Add back the old columns
ALTER TABLE subscriptions 
    ADD COLUMN IF NOT EXISTS original_price DECIMAL(10, 2),
    ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS final_price DECIMAL(10, 2),
    ADD COLUMN IF NOT EXISTS voucher_id UUID REFERENCES vouchers(id),
    ADD COLUMN IF NOT EXISTS voucher_code VARCHAR(50);

-- Migrate data: set original_price and final_price from price
UPDATE subscriptions 
SET 
    original_price = price,
    final_price = price,
    discount_amount = 0
WHERE original_price IS NULL OR final_price IS NULL;

-- Make columns NOT NULL after data migration
ALTER TABLE subscriptions 
    ALTER COLUMN original_price SET NOT NULL,
    ALTER COLUMN final_price SET NOT NULL;

-- Drop the price column
ALTER TABLE subscriptions 
    DROP COLUMN IF EXISTS price;

-- Recreate constraints with old values
ALTER TABLE subscriptions 
    ADD CONSTRAINT subscriptions_access_type_check 
    CHECK (access_type IN ('paid', 'free_mf', 'free_voucher', 'admin_granted'));

ALTER TABLE subscriptions 
    ADD CONSTRAINT subscriptions_price_check 
    CHECK (final_price >= 0);

