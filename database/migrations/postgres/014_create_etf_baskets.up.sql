-- Migration: Create ETF Baskets Tables
-- Description: Creates etf_baskets and etf_basket_items tables
-- Created: 2025-01-XX

CREATE TABLE etf_baskets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advisory_type_id UUID NOT NULL REFERENCES advisory_types(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    report_url TEXT,
    report_file_name VARCHAR(255),
    report_uploaded_at TIMESTAMP,
    report_uploaded_by UUID REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    published_by UUID REFERENCES users(id),
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT etf_baskets_status_check CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE TABLE etf_basket_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    etf_basket_id UUID NOT NULL REFERENCES etf_baskets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(50),
    cmp DECIMAL(10, 2) NOT NULL,
    target DECIMAL(10, 2) NOT NULL,
    stop_loss DECIMAL(10, 2),
    entry_range_min DECIMAL(10, 2),
    entry_range_max DECIMAL(10, 2),
    action VARCHAR(10),
    risk_level VARCHAR(50),
    time_horizon VARCHAR(100),
    rationale TEXT,
    pdf_link TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_price DECIMAL(10, 2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT etf_basket_items_action_check CHECK (action IN ('buy', 'sell')),
    CONSTRAINT etf_basket_items_status_check CHECK (status IN ('active', 'target_achieved', 'stop_loss_hit', 'closed')),
    CONSTRAINT etf_basket_items_entry_range_check CHECK (entry_range_max >= entry_range_min OR entry_range_max IS NULL OR entry_range_min IS NULL),
    CONSTRAINT etf_basket_items_price_check CHECK (cmp > 0 AND target > 0)
);

CREATE INDEX idx_etf_baskets_advisory_type_id ON etf_baskets(advisory_type_id);
CREATE INDEX idx_etf_baskets_status ON etf_baskets(status);
CREATE INDEX idx_etf_baskets_published_at ON etf_baskets(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_etf_baskets_published_by ON etf_baskets(published_by);

CREATE INDEX idx_etf_basket_items_etf_basket_id ON etf_basket_items(etf_basket_id);
CREATE INDEX idx_etf_basket_items_symbol ON etf_basket_items(symbol) WHERE symbol IS NOT NULL;
CREATE INDEX idx_etf_basket_items_status ON etf_basket_items(status);
CREATE INDEX idx_etf_basket_items_display_order ON etf_basket_items(etf_basket_id, display_order);
