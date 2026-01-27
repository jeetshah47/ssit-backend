-- Migration: Rollback Seed Dummy Data
-- Description: Removes all dummy data inserted by seed migration

-- Delete Weekly Audio
DELETE FROM weekly_audio WHERE title IN (
    'Understanding the FII sell-off cycle',
    'Banking sector earnings analysis',
    'Small-cap valuation concerns'
);

-- Delete Weekly Market Mood
DELETE FROM weekly_market_mood WHERE title IN (
    'Steady consolidation amidst global volatility',
    'Market resilience in face of global headwinds',
    'Consolidation phase before next leg up'
);

-- Delete Webinars
DELETE FROM webinars WHERE title IN (
    'The manufacturing renaissance in India',
    'Post-Election market dynamics',
    'Cracking the Small-cap puzzle',
    'Banking Sector Outlook 2025',
    'FII Flow Analysis'
);

-- Delete NFOs
DELETE FROM nfos WHERE name IN (
    'Motilal Oswal Business Cycle Fund',
    'HDFC Manufacturing Fund',
    'ICICI Prudential Infrastructure Fund'
);

-- Delete MF Schemes
DELETE FROM mf_schemes WHERE name IN (
    'HDFC Top 100 Fund',
    'SBI Bluechip Fund',
    'ICICI Prudential Value Discovery',
    'Axis Midcap Fund',
    'Kotak Emerging Equity Fund'
);

-- Delete Mutual Fund Basket Items
DELETE FROM mutual_fund_basket_items WHERE scheme_name IN (
    'HDFC Top 100 Fund',
    'SBI Bluechip Fund',
    'ICICI Prudential Value Discovery',
    'SBI Banking & Financial Services Fund',
    'ICICI Prudential Banking & Financial Services'
);

-- Delete Mutual Fund Baskets
DELETE FROM mutual_fund_baskets WHERE name IN (
    'Top Equity Funds',
    'Banking Sector Funds'
);

-- Delete ETF Basket Items
DELETE FROM etf_basket_items WHERE name IN (
    'Nippon India Nifty 50 ETF',
    'HDFC Nifty 50 ETF',
    'Banking ETF'
);

-- Delete ETF Baskets
DELETE FROM etf_baskets WHERE name IN (
    'Nifty 50 ETF Basket',
    'Sectoral ETF Picks'
);

-- Delete IPO Advisories
DELETE FROM ipo_advisories WHERE ipo_name IN (
    'Motilal Oswal Business Cycle Fund',
    'Tech Mahindra Digital',
    'Green Energy Solutions'
);

-- Delete Stock Basket Items
DELETE FROM stock_basket_items WHERE stock_name IN (
    'Reliance Industries',
    'TCS',
    'HDFC Bank',
    'Tata Motors',
    'Adani Ports',
    'Bajaj Finance'
);

-- Delete Stock Baskets
DELETE FROM stock_baskets WHERE name IN (
    'Large Cap Opportunities',
    'Mid Cap Growth Picks'
);

-- Reset Sector Snapshots to default values
UPDATE sector_snapshots SET 
    current_value = NULL,
    change_percentage = NULL,
    change_value = NULL,
    last_updated = NULL
WHERE sector_name IN (
    'Banking & Financials',
    'Automobiles',
    'Information Technology',
    'Pharmaceuticals',
    'FMCG'
);

DELETE FROM sector_snapshots WHERE sector_name IN (
    'Banking & Financials',
    'Automobiles',
    'Information Technology',
    'Pharmaceuticals',
    'FMCG'
);

-- Reset default sectors
UPDATE sector_snapshots SET 
    current_value = NULL,
    change_percentage = NULL,
    change_value = NULL,
    last_updated = NULL
WHERE sector_name IN ('NIFTY', 'BANK NIFTY', 'SENSEX');
