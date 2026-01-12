-- Rollback: Revert Payment Method Constraint Fix
-- Description: Reverts the payment_method constraint

-- Drop the existing constraint
ALTER TABLE payments 
    DROP CONSTRAINT IF EXISTS payments_method_check;

-- Restore constraint without 'paytm' (back to migration 010 state)
ALTER TABLE payments 
    ADD CONSTRAINT payments_method_check 
    CHECK (payment_method IN ('upi', 'card', 'netbanking', 'razorpay', 'paytm'));

