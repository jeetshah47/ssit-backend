-- Migration: Update Payment Method Constraint
-- Description: Updates the payment_method constraint to allow 'razorpay' as a valid value
--              This is needed because payment_method can represent the gateway when the actual method is not yet known

-- Drop the existing constraint
ALTER TABLE payments 
    DROP CONSTRAINT IF EXISTS payments_method_check;

-- Add new constraint that includes 'razorpay' as a valid payment method
-- This allows us to set payment_method to 'razorpay' initially, and update it later
-- when we know the actual method chosen by the user (UPI, card, netbanking)
ALTER TABLE payments 
    ADD CONSTRAINT payments_method_check 
    CHECK (payment_method IN ('upi', 'card', 'netbanking', 'razorpay'));

