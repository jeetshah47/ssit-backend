-- Migration: Rollback Update Payment Method Constraint
-- Description: Restores the original payment_method constraint

-- Drop the updated constraint
ALTER TABLE payments 
    DROP CONSTRAINT IF EXISTS payments_method_check;

-- Restore original constraint (only upi, card, netbanking)
ALTER TABLE payments 
    ADD CONSTRAINT payments_method_check 
    CHECK (payment_method IN ('upi', 'card', 'netbanking'));

