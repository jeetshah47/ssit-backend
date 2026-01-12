-- Rollback: Remove Paytm from Payment Method Constraint
-- Description: Reverts the payment_method constraint to exclude 'paytm'

-- Drop the existing constraint
ALTER TABLE payments 
    DROP CONSTRAINT IF EXISTS payments_method_check;

-- Restore constraint without 'paytm'
ALTER TABLE payments 
    ADD CONSTRAINT payments_method_check 
    CHECK (payment_method IN ('upi', 'card', 'netbanking', 'razorpay'));

