-- Migration: Create Stock Baskets Tables
-- Description: Creates stock_baskets and stock_basket_items tables
-- Created: 2025-01-XX

CREATE TABLE stock_baskets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advisory_type_id UUID NOT NULL REFERENCES advisory_types(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_bullet_idea BOOLEAN NOT NULL DEFAULT false,
    report_url TEXT,
    report_file_name VARCHAR(255),
    report_uploaded_at TIMESTAMP,
    report_uploaded_by UUID REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    published_by UUID REFERENCES users(id),
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT stock_baskets_status_check CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE TABLE stock_basket_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_basket_id UUID NOT NULL REFERENCES stock_baskets(id) ON DELETE CASCADE,
    stock_name VARCHAR(255) NOT NULL,
    stock_symbol VARCHAR(50),
    cmp DECIMAL(10, 2) NOT NULL,
    target DECIMAL(10, 2) NOT NULL,
    stop_loss DECIMAL(10, 2),
    entry_range_min DECIMAL(10, 2),
    entry_range_max DECIMAL(10, 2),
    action VARCHAR(10),
    risk_level VARCHAR(50),
    time_horizon VARCHAR(100),
    rationale TEXT,
    report_url TEXT,
    fundamentals_url TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_bullet_idea BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_price DECIMAL(10, 2),
    target_achieved_at TIMESTAMP,
    stop_loss_hit_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT stock_basket_items_action_check CHECK (action IN ('buy', 'sell')),
    CONSTRAINT stock_basket_items_status_check CHECK (status IN ('active', 'target_achieved', 'stop_loss_hit', 'closed')),
    CONSTRAINT stock_basket_items_entry_range_check CHECK (entry_range_max >= entry_range_min OR entry_range_max IS NULL OR entry_range_min IS NULL),
    CONSTRAINT stock_basket_items_price_check CHECK (cmp > 0 AND target > 0)
);

CREATE INDEX idx_stock_baskets_advisory_type_id ON stock_baskets(advisory_type_id);
CREATE INDEX idx_stock_baskets_status ON stock_baskets(status);
CREATE INDEX idx_stock_baskets_published_at ON stock_baskets(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_stock_baskets_published_by ON stock_baskets(published_by);
CREATE INDEX idx_stock_baskets_is_bullet_idea ON stock_baskets(is_bullet_idea) WHERE is_bullet_idea = true;

CREATE INDEX idx_stock_basket_items_stock_basket_id ON stock_basket_items(stock_basket_id);
CREATE INDEX idx_stock_basket_items_stock_symbol ON stock_basket_items(stock_symbol) WHERE stock_symbol IS NOT NULL;
CREATE INDEX idx_stock_basket_items_status ON stock_basket_items(status);
CREATE INDEX idx_stock_basket_items_display_order ON stock_basket_items(stock_basket_id, display_order);
CREATE INDEX idx_stock_basket_items_is_bullet_idea ON stock_basket_items(is_bullet_idea) WHERE is_bullet_idea = true;
