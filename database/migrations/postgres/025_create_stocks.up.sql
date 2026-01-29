-- Migration: Create stocks table (master stock list for bullets/recommendations)
-- Description: Master list of stocks; admin adds here, then selects from list for bullets and recommendations

CREATE TABLE stocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(50),
    exchange VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stocks_name ON stocks(name);
CREATE INDEX idx_stocks_symbol ON stocks(symbol) WHERE symbol IS NOT NULL;
CREATE INDEX idx_stocks_exchange ON stocks(exchange) WHERE exchange IS NOT NULL;
