-- Migration: Create Mutual Fund Baskets Tables
-- Description: Creates mutual_fund_baskets and mutual_fund_basket_items tables
-- Created: 2025-01-XX

CREATE TABLE mutual_fund_baskets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advisory_type_id UUID NOT NULL REFERENCES advisory_types(id),
    basket_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    max_schemes INTEGER NOT NULL DEFAULT 10,
    report_url TEXT,
    report_file_name VARCHAR(255),
    report_uploaded_at TIMESTAMP,
    report_uploaded_by UUID REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    published_by UUID REFERENCES users(id),
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT mutual_fund_baskets_basket_type_check CHECK (basket_type IN ('general', 'index_fund', 'sectoral', 'nfo_review')),
    CONSTRAINT mutual_fund_baskets_status_check CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT mutual_fund_baskets_max_schemes_check CHECK (max_schemes > 0 AND max_schemes <= 10)
);

CREATE TABLE mutual_fund_basket_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mutual_fund_basket_id UUID NOT NULL REFERENCES mutual_fund_baskets(id) ON DELETE CASCADE,
    scheme_name VARCHAR(255) NOT NULL,
    scheme_code VARCHAR(50),
    entry_price DECIMAL(10, 2) NOT NULL,
    exit_price DECIMAL(10, 2),
    current_nav DECIMAL(10, 4),
    trend VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT mutual_fund_basket_items_trend_check CHECK (trend IN ('positive', 'negative', 'neutral')),
    CONSTRAINT mutual_fund_basket_items_status_check CHECK (status IN ('active', 'target_achieved', 'closed')),
    CONSTRAINT mutual_fund_basket_items_entry_price_check CHECK (entry_price > 0)
);

CREATE INDEX idx_mutual_fund_baskets_advisory_type_id ON mutual_fund_baskets(advisory_type_id);
CREATE INDEX idx_mutual_fund_baskets_basket_type ON mutual_fund_baskets(basket_type);
CREATE INDEX idx_mutual_fund_baskets_status ON mutual_fund_baskets(status);
CREATE INDEX idx_mutual_fund_baskets_published_at ON mutual_fund_baskets(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_mutual_fund_baskets_published_by ON mutual_fund_baskets(published_by);

CREATE INDEX idx_mutual_fund_basket_items_mutual_fund_basket_id ON mutual_fund_basket_items(mutual_fund_basket_id);
CREATE INDEX idx_mutual_fund_basket_items_scheme_code ON mutual_fund_basket_items(scheme_code) WHERE scheme_code IS NOT NULL;
CREATE INDEX idx_mutual_fund_basket_items_status ON mutual_fund_basket_items(status);
CREATE INDEX idx_mutual_fund_basket_items_display_order ON mutual_fund_basket_items(mutual_fund_basket_id, display_order);
