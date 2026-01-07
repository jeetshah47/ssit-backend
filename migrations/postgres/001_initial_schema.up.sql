-- Migration: Initial Schema
-- Description: Creates all core tables for the Equitywala Stock Advisory Platform
-- Created: 2025-01-XX

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- AUTH MODULE
-- ============================================

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20),
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- User Classification (Critical for MF vs Paid logic)
    is_mf_customer BOOLEAN NOT NULL DEFAULT false,
    mf_customer_id VARCHAR(100),
    classification_source VARCHAR(50),
    classification_notes TEXT,
    
    -- Account Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending_verification',
    
    -- Email Verification
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMP,
    
    -- Profile Settings
    username VARCHAR(100),
    website VARCHAR(255),
    bio TEXT,
    job_title VARCHAR(255),
    show_job_title BOOLEAN DEFAULT false,
    alternative_email VARCHAR(255),
    profile_photo_url TEXT,
    cover_photo_url TEXT,
    
    -- OAuth
    google_id VARCHAR(255) UNIQUE,
    oauth_provider VARCHAR(50),
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT users_status_check CHECK (status IN ('pending_verification', 'verified', 'active', 'suspended', 'deleted'))
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_is_mf_customer ON users(is_mf_customer);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE UNIQUE INDEX idx_users_phone_unique ON users(phone) WHERE phone IS NOT NULL;
CREATE UNIQUE INDEX idx_users_mf_customer_id_unique ON users(mf_customer_id) WHERE mf_customer_id IS NOT NULL;

-- OTPs table
CREATE TABLE otps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    used_at TIMESTAMP,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT otps_type_check CHECK (type IN ('email_verification', 'password_reset', 'login'))
);

CREATE INDEX idx_otps_user_id ON otps(user_id);
CREATE INDEX idx_otps_email_type ON otps(email, type);
CREATE INDEX idx_otps_expires_at ON otps(expires_at) WHERE used = false;

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    refresh_token_hash VARCHAR(255),
    
    -- Device Information
    device VARCHAR(255),
    device_type VARCHAR(50),
    user_agent TEXT,
    ip_address VARCHAR(45),
    location VARCHAR(255),
    
    -- Session Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    
    CONSTRAINT sessions_device_type_check CHECK (device_type IN ('desktop', 'mobile', 'tablet'))
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_user_active ON sessions(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at) WHERE is_active = true;

-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_system_role BOOLEAN NOT NULL DEFAULT false,
    permissions JSONB NOT NULL DEFAULT '[]',
    member_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT roles_name_check CHECK (name ~ '^[A-Za-z0-9_]+$')
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_is_system_role ON roles(is_system_role);

-- User roles table
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT user_roles_unique UNIQUE (user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- ============================================
-- KYC MODULE
-- ============================================

-- KYC records table
CREATE TABLE kyc_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- KYC Information
    pan VARCHAR(10) NOT NULL,
    pan_verified BOOLEAN NOT NULL DEFAULT false,
    
    -- Document URLs (stored in S3)
    pan_card_url TEXT,
    aadhaar_url TEXT,
    bank_statement_url TEXT,
    
    -- Verification Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Provider Information
    provider VARCHAR(100),
    provider_reference_id VARCHAR(255),
    provider_response JSONB,
    
    -- Verification Details
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMP,
    rejection_reason TEXT,
    
    -- Metadata
    submitted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT kyc_status_check CHECK (status IN ('pending', 'submitted', 'under_review', 'verified', 'failed', 'rejected')),
    CONSTRAINT kyc_pan_format CHECK (pan ~ '^[A-Z]{5}[0-9]{4}[A-Z]{1}$')
);

CREATE INDEX idx_kyc_user_id ON kyc_records(user_id);
CREATE INDEX idx_kyc_status ON kyc_records(status);
CREATE INDEX idx_kyc_pan ON kyc_records(pan);
CREATE INDEX idx_kyc_provider_reference ON kyc_records(provider_reference_id) WHERE provider_reference_id IS NOT NULL;
CREATE INDEX idx_kyc_created_at ON kyc_records(created_at);

-- ============================================
-- SUBSCRIPTION MODULE
-- ============================================

-- Pricing packages table
CREATE TABLE pricing_packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Pricing
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    
    -- Duration
    duration_days INTEGER NOT NULL,
    duration_type VARCHAR(50) NOT NULL,
    
    -- Access Scope
    access_level VARCHAR(50) NOT NULL DEFAULT 'standard',
    features JSONB DEFAULT '[]',
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    is_published BOOLEAN NOT NULL DEFAULT false,
    
    -- Metadata
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT pricing_duration_type_check CHECK (duration_type IN ('monthly', 'quarterly', 'annual', 'custom')),
    CONSTRAINT pricing_status_check CHECK (status IN ('active', 'inactive', 'archived')),
    CONSTRAINT pricing_price_check CHECK (price >= 0)
);

CREATE INDEX idx_pricing_status ON pricing_packages(status);
CREATE INDEX idx_pricing_is_published ON pricing_packages(is_published) WHERE is_published = true;
CREATE INDEX idx_pricing_created_at ON pricing_packages(created_at);

-- Vouchers table
CREATE TABLE vouchers (
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

CREATE INDEX idx_vouchers_code ON vouchers(code);
CREATE INDEX idx_vouchers_active ON vouchers(is_active) WHERE is_active = true;
CREATE INDEX idx_vouchers_validity ON vouchers(valid_from, valid_until);

-- Subscriptions table
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_id UUID NOT NULL REFERENCES pricing_packages(id),
    
    -- Pricing
    original_price DECIMAL(10, 2) NOT NULL,
    discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    final_price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    
    -- Voucher
    voucher_id UUID REFERENCES vouchers(id),
    voucher_code VARCHAR(50),
    
    -- Access Type (Critical for MF customer logic)
    access_type VARCHAR(50) NOT NULL,
    access_reason TEXT,
    
    -- Validity
    starts_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    -- Payment Reference
    payment_id UUID,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP,
    
    CONSTRAINT subscriptions_access_type_check CHECK (access_type IN ('paid', 'free_mf', 'free_voucher', 'admin_granted')),
    CONSTRAINT subscriptions_status_check CHECK (status IN ('pending', 'active', 'expired', 'cancelled', 'refunded')),
    CONSTRAINT subscriptions_validity_check CHECK (expires_at > starts_at),
    CONSTRAINT subscriptions_price_check CHECK (final_price >= 0)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_package_id ON subscriptions(package_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_active ON subscriptions(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_subscriptions_expires_at ON subscriptions(expires_at);
CREATE INDEX idx_subscriptions_access_type ON subscriptions(access_type);

-- Voucher redemptions table
CREATE TABLE voucher_redemptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    voucher_id UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    
    discount_amount DECIMAL(10, 2) NOT NULL,
    redeemed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT voucher_redemptions_unique UNIQUE (voucher_id, user_id, subscription_id)
);

CREATE INDEX idx_voucher_redemptions_voucher_id ON voucher_redemptions(voucher_id);
CREATE INDEX idx_voucher_redemptions_user_id ON voucher_redemptions(user_id);
CREATE INDEX idx_voucher_redemptions_redeemed_at ON voucher_redemptions(redeemed_at);

-- ============================================
-- PAYMENT MODULE
-- ============================================

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id),
    
    -- Amount
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    
    -- Payment Method
    payment_method VARCHAR(50) NOT NULL,
    payment_gateway VARCHAR(100),
    
    -- Gateway Information
    gateway_transaction_id VARCHAR(255),
    gateway_order_id VARCHAR(255),
    gateway_payment_id VARCHAR(255),
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Payment Details
    payment_data JSONB,
    failure_reason TEXT,
    
    -- Refund Information
    refund_amount DECIMAL(10, 2),
    refund_transaction_id VARCHAR(255),
    refunded_by UUID REFERENCES users(id),
    refunded_at TIMESTAMP,
    refund_reason TEXT,
    
    -- Idempotency
    idempotency_key VARCHAR(255) UNIQUE,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    
    CONSTRAINT payments_status_check CHECK (status IN ('pending', 'processing', 'success', 'failed', 'refunded', 'cancelled')),
    CONSTRAINT payments_method_check CHECK (payment_method IN ('upi', 'card', 'netbanking')),
    CONSTRAINT payments_amount_check CHECK (amount > 0),
    CONSTRAINT payments_refund_check CHECK (refund_amount IS NULL OR (refund_amount >= 0 AND refund_amount <= amount))
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_subscription_id ON payments(subscription_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_gateway_transaction_id ON payments(gateway_transaction_id) WHERE gateway_transaction_id IS NOT NULL;
CREATE INDEX idx_payments_idempotency_key ON payments(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_payments_created_at ON payments(created_at);

-- Payment webhooks table
CREATE TABLE payment_webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID REFERENCES payments(id),
    
    -- Webhook Data
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    signature VARCHAR(255),
    
    -- Status
    processed BOOLEAN NOT NULL DEFAULT false,
    processed_at TIMESTAMP,
    error_message TEXT,
    
    -- Metadata
    received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_webhooks_payment_id ON payment_webhooks(payment_id);
CREATE INDEX idx_payment_webhooks_processed ON payment_webhooks(processed) WHERE processed = false;
CREATE INDEX idx_payment_webhooks_received_at ON payment_webhooks(received_at);

-- ============================================
-- ADVISORY MODULE
-- ============================================

-- Advisories table
CREATE TABLE advisories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Stock Information
    stock_name VARCHAR(255) NOT NULL,
    stock_symbol VARCHAR(50) NOT NULL,
    
    -- Recommendation
    action VARCHAR(10) NOT NULL,
    entry_range_min DECIMAL(10, 2) NOT NULL,
    entry_range_max DECIMAL(10, 2) NOT NULL,
    target DECIMAL(10, 2) NOT NULL,
    stop_loss DECIMAL(10, 2) NOT NULL,
    time_horizon VARCHAR(100),
    rationale TEXT NOT NULL,
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    -- Pricing
    current_price DECIMAL(10, 2),
    target_achieved_at TIMESTAMP,
    stop_loss_hit_at TIMESTAMP,
    closed_at TIMESTAMP,
    
    -- Publishing
    published_by UUID NOT NULL REFERENCES users(id),
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Access Control
    access_level VARCHAR(50) NOT NULL DEFAULT 'standard',
    
    -- Metadata (Immutable after publish)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMP,
    
    CONSTRAINT advisories_action_check CHECK (action IN ('buy', 'sell')),
    CONSTRAINT advisories_status_check CHECK (status IN ('active', 'target_achieved', 'stop_loss_hit', 'closed', 'archived')),
    CONSTRAINT advisories_entry_range_check CHECK (entry_range_max >= entry_range_min),
    CONSTRAINT advisories_price_check CHECK (
        entry_range_min > 0 AND entry_range_max > 0 AND 
        target > 0 AND stop_loss > 0
    )
);

CREATE INDEX idx_advisories_stock_symbol ON advisories(stock_symbol);
CREATE INDEX idx_advisories_status ON advisories(status);
CREATE INDEX idx_advisories_published_at ON advisories(published_at DESC);
CREATE INDEX idx_advisories_access_level ON advisories(access_level);
CREATE INDEX idx_advisories_published_by ON advisories(published_by);

-- Advisory views table
CREATE TABLE advisory_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advisory_id UUID NOT NULL REFERENCES advisories(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT advisory_views_unique UNIQUE (advisory_id, user_id)
);

CREATE INDEX idx_advisory_views_advisory_id ON advisory_views(advisory_id);
CREATE INDEX idx_advisory_views_user_id ON advisory_views(user_id);
CREATE INDEX idx_advisory_views_viewed_at ON advisory_views(viewed_at);

-- ============================================
-- SETTINGS MODULE
-- ============================================

-- Notification settings table
CREATE TABLE notification_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    
    -- Email Settings
    email_news_and_updates BOOLEAN NOT NULL DEFAULT true,
    email_tips_and_tutorials BOOLEAN NOT NULL DEFAULT true,
    email_user_research BOOLEAN NOT NULL DEFAULT false,
    email_comments VARCHAR(50) NOT NULL DEFAULT 'all-comments',
    email_reminders VARCHAR(50) NOT NULL DEFAULT 'all-reminders',
    email_more_activity VARCHAR(50) NOT NULL DEFAULT 'all-activity',
    
    -- In-App Settings
    in_app_news_and_updates BOOLEAN NOT NULL DEFAULT true,
    in_app_tips_and_tutorials BOOLEAN NOT NULL DEFAULT true,
    in_app_user_research BOOLEAN NOT NULL DEFAULT false,
    in_app_comments VARCHAR(50) NOT NULL DEFAULT 'all-comments',
    in_app_reminders VARCHAR(50) NOT NULL DEFAULT 'all-reminders',
    in_app_more_activity VARCHAR(50) NOT NULL DEFAULT 'all-activity',
    
    -- Push Settings
    push_news_and_updates BOOLEAN NOT NULL DEFAULT true,
    push_tips_and_tutorials BOOLEAN NOT NULL DEFAULT true,
    push_user_research BOOLEAN NOT NULL DEFAULT false,
    push_comments VARCHAR(50) NOT NULL DEFAULT 'all-comments',
    push_reminders VARCHAR(50) NOT NULL DEFAULT 'all-reminders',
    push_more_activity VARCHAR(50) NOT NULL DEFAULT 'all-activity',
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT notification_email_comments_check CHECK (email_comments IN ('do-not-notify', 'mentions-only', 'all-comments')),
    CONSTRAINT notification_email_reminders_check CHECK (email_reminders IN ('do-not-notify', 'important-only', 'all-reminders')),
    CONSTRAINT notification_email_activity_check CHECK (email_more_activity IN ('do-not-notify', 'all-activity')),
    CONSTRAINT notification_in_app_comments_check CHECK (in_app_comments IN ('do-not-notify', 'mentions-only', 'all-comments')),
    CONSTRAINT notification_in_app_reminders_check CHECK (in_app_reminders IN ('do-not-notify', 'important-only', 'all-reminders')),
    CONSTRAINT notification_in_app_activity_check CHECK (in_app_more_activity IN ('do-not-notify', 'all-activity')),
    CONSTRAINT notification_push_comments_check CHECK (push_comments IN ('do-not-notify', 'mentions-only', 'all-comments')),
    CONSTRAINT notification_push_reminders_check CHECK (push_reminders IN ('do-not-notify', 'important-only', 'all-reminders')),
    CONSTRAINT notification_push_activity_check CHECK (push_more_activity IN ('do-not-notify', 'all-activity'))
);

-- Team members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id),
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Invitation
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invite_token VARCHAR(255) UNIQUE,
    invite_expires_at TIMESTAMP,
    joined_at TIMESTAMP,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    removed_at TIMESTAMP,
    
    CONSTRAINT team_members_status_check CHECK (status IN ('pending', 'active', 'inactive')),
    CONSTRAINT team_members_unique UNIQUE (team_id, user_id) WHERE team_id IS NOT NULL
);

CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_role_id ON team_members(role_id);
CREATE INDEX idx_team_members_status ON team_members(status);
CREATE INDEX idx_team_members_invite_token ON team_members(invite_token) WHERE invite_token IS NOT NULL;

-- ============================================
-- ADMIN MODULE
-- ============================================

-- User classification logs table
CREATE TABLE user_classification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Classification Change
    old_classification BOOLEAN,
    new_classification BOOLEAN NOT NULL,
    classification_source VARCHAR(50) NOT NULL,
    
    -- Change Details
    changed_by UUID NOT NULL REFERENCES users(id),
    reason TEXT NOT NULL,
    mf_customer_id VARCHAR(100),
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_classification_logs_user_id ON user_classification_logs(user_id);
CREATE INDEX idx_user_classification_logs_changed_by ON user_classification_logs(changed_by);
CREATE INDEX idx_user_classification_logs_created_at ON user_classification_logs(created_at);

-- ============================================
-- TRIGGERS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kyc_records_updated_at BEFORE UPDATE ON kyc_records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pricing_packages_updated_at BEFORE UPDATE ON pricing_packages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vouchers_updated_at BEFORE UPDATE ON vouchers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_advisories_updated_at BEFORE UPDATE ON advisories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_settings_updated_at BEFORE UPDATE ON notification_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_team_members_updated_at BEFORE UPDATE ON team_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

