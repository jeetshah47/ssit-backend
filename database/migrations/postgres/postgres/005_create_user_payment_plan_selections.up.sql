-- Migration: Create User Payment Plan Selections Table
-- Description: Creates normalized table for user payment plan selections during signup
-- Created: 2025-01-XX

-- Create user_payment_plan_selections table
CREATE TABLE user_payment_plan_selections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_id UUID NOT NULL REFERENCES pricing_packages(id),
    
    -- Selection metadata
    selected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- Values: 'pending', 'confirmed', 'cancelled'
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT user_payment_plan_selections_status_check CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    CONSTRAINT user_payment_plan_selections_unique_user UNIQUE (user_id) -- One selection per user
);

CREATE INDEX idx_user_payment_plan_selections_user_id ON user_payment_plan_selections(user_id);
CREATE INDEX idx_user_payment_plan_selections_package_id ON user_payment_plan_selections(package_id);
CREATE INDEX idx_user_payment_plan_selections_status ON user_payment_plan_selections(status);

