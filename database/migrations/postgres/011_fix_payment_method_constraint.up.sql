-- Migration: Fix Payment Method Constraint for Paytm
-- Description: Ensures 'paytm' is included in the payment_method constraint
--              This is a fix in case migration 010 didn't apply correctly

-- Drop the existing constraint
ALTER TABLE payments 
    DROP CONSTRAINT IF EXISTS payments_method_check;

-- Add new constraint that includes 'paytm' as a valid payment method
ALTER TABLE payments 
    ADD CONSTRAINT payments_method_check 
    CHECK (payment_method IN ('upi', 'card', 'netbanking', 'razorpay', 'paytm'));

