-- Migration: Seed Pricing Packages
-- Description: Inserts initial pricing packages for Standard, Plus, and Premium plans
--              with quarterly and annual billing periods

-- Standard Plan - Quarterly
INSERT INTO pricing_packages (id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at, published_at)
VALUES (
    gen_random_uuid(),
    'Standard',
    'Standard plan with quarterly billing - Includes basic features for mutual funds, equities, and tech support',
    5000.00,
    'INR',
    90, -- 3 months (quarterly)
    'quarterly',
    'standard',
    'active',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Standard Plan - Annual
INSERT INTO pricing_packages (id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at, published_at)
VALUES (
    gen_random_uuid(),
    'Standard',
    'Standard plan with annual billing - Includes basic features for mutual funds, equities, and tech support',
    18000.00,
    'INR',
    365, -- 1 year (annual)
    'annual',
    'standard',
    'active',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Plus Plan - Quarterly
INSERT INTO pricing_packages (id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at, published_at)
VALUES (
    gen_random_uuid(),
    'Plus',
    'Plus plan with quarterly billing - Includes advanced features with sectoral baskets, weekly market mood, and live webinars',
    14000.00,
    'INR',
    90, -- 3 months (quarterly)
    'quarterly',
    'plus',
    'active',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Plus Plan - Annual
INSERT INTO pricing_packages (id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at, published_at)
VALUES (
    gen_random_uuid(),
    'Plus',
    'Plus plan with annual billing - Includes advanced features with sectoral baskets, weekly market mood, and live webinars',
    50000.00,
    'INR',
    365, -- 1 year (annual)
    'annual',
    'plus',
    'active',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Premium Plan - Quarterly
INSERT INTO pricing_packages (id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at, published_at)
VALUES (
    gen_random_uuid(),
    'Premium',
    'Premium plan with quarterly billing - Includes all features with detailed reports, commodities, PMS/AIF, and in-person analyst meetings',
    35000.00,
    'INR',
    90, -- 3 months (quarterly)
    'quarterly',
    'premium',
    'active',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Premium Plan - Annual
INSERT INTO pricing_packages (id, name, description, price, currency, duration_days, duration_type, access_level, status, is_published, created_at, updated_at, published_at)
VALUES (
    gen_random_uuid(),
    'Premium',
    'Premium plan with annual billing - Includes all features with detailed reports, commodities, PMS/AIF, and in-person analyst meetings',
    140000.00,
    'INR',
    365, -- 1 year (annual)
    'annual',
    'premium',
    'active',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

