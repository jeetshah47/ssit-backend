-- Migration: Seed Dummy Data for Frontend
-- Description: Inserts dummy data for all frontend pages (Overview, Equities, Mutual Funds, Data Room)
-- Created: 2025-01-XX

-- Ensure UUID extension is enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Get advisory type IDs (assuming they exist from migration 012)
DO $$
DECLARE
    stock_basket_type_id UUID;
    etf_basket_type_id UUID;
    ipo_type_id UUID;
    mf_basket_type_id UUID;
    dummy_user_id UUID := uuid_generate_v4();
BEGIN
    -- Get advisory type IDs
    SELECT id INTO stock_basket_type_id FROM advisory_types WHERE name = 'stock_basket' LIMIT 1;
    SELECT id INTO etf_basket_type_id FROM advisory_types WHERE name = 'etf_basket' LIMIT 1;
    SELECT id INTO ipo_type_id FROM advisory_types WHERE name = 'ipo' LIMIT 1;
    SELECT id INTO mf_basket_type_id FROM advisory_types WHERE name = 'mutual_fund_basket' LIMIT 1;

    -- If advisory types don't exist, create them
    IF stock_basket_type_id IS NULL THEN
        INSERT INTO advisory_types (id, name, display_name, description) 
        VALUES (uuid_generate_v4(), 'stock_basket', 'Stock Basket', 'Basket of stock recommendations')
        RETURNING id INTO stock_basket_type_id;
    END IF;
    
    IF etf_basket_type_id IS NULL THEN
        INSERT INTO advisory_types (id, name, display_name, description) 
        VALUES (uuid_generate_v4(), 'etf_basket', 'ETF Basket', 'Basket of ETF recommendations')
        RETURNING id INTO etf_basket_type_id;
    END IF;
    
    IF ipo_type_id IS NULL THEN
        INSERT INTO advisory_types (id, name, display_name, description) 
        VALUES (uuid_generate_v4(), 'ipo', 'IPO Advisory', 'Initial Public Offering advisory')
        RETURNING id INTO ipo_type_id;
    END IF;
    
    IF mf_basket_type_id IS NULL THEN
        INSERT INTO advisory_types (id, name, display_name, description) 
        VALUES (uuid_generate_v4(), 'mutual_fund_basket', 'Mutual Fund Basket', 'Basket of mutual fund recommendations')
        RETURNING id INTO mf_basket_type_id;
    END IF;

    -- ============================================
    -- 1. SECTOR SNAPSHOTS (Overview & Data Room)
    -- ============================================
    UPDATE sector_snapshots SET 
        current_value = 24050.75,
        change_percentage = 0.85,
        change_value = 203.25,
        last_updated = CURRENT_TIMESTAMP
    WHERE sector_name = 'NIFTY';

    UPDATE sector_snapshots SET 
        current_value = 51234.50,
        change_percentage = 1.25,
        change_value = 632.75,
        last_updated = CURRENT_TIMESTAMP
    WHERE sector_name = 'BANK NIFTY';

    UPDATE sector_snapshots SET 
        current_value = 80123.45,
        change_percentage = 0.65,
        change_value = 520.30,
        last_updated = CURRENT_TIMESTAMP
    WHERE sector_name = 'SENSEX';

    -- Additional sectors
    INSERT INTO sector_snapshots (sector_name, current_value, change_percentage, change_value, last_updated)
    VALUES 
        ('Banking & Financials', 24567.89, 1.15, 280.45, CURRENT_TIMESTAMP),
        ('Automobiles', 18923.45, -0.35, -66.20, CURRENT_TIMESTAMP),
        ('Information Technology', 32145.67, 0.95, 305.20, CURRENT_TIMESTAMP),
        ('Pharmaceuticals', 15678.90, 0.45, 70.55, CURRENT_TIMESTAMP),
        ('FMCG', 22345.67, 0.75, 166.60, CURRENT_TIMESTAMP)
    ON CONFLICT (sector_name) DO UPDATE SET
        current_value = EXCLUDED.current_value,
        change_percentage = EXCLUDED.change_percentage,
        change_value = EXCLUDED.change_value,
        last_updated = EXCLUDED.last_updated;

    -- ============================================
    -- 2. STOCK BASKETS (Overview & Equities)
    -- ============================================
    -- Stock Basket 1: Large Cap Opportunities
    INSERT INTO stock_baskets (advisory_type_id, name, description, is_bullet_idea, status, published_at)
    VALUES (
        stock_basket_type_id,
        'Large Cap Opportunities',
        'Top large-cap stocks with strong fundamentals and growth potential',
        false,
        'published',
        CURRENT_TIMESTAMP
    );

    -- Stock Basket 2: Mid Cap Growth (Bullet Idea)
    INSERT INTO stock_baskets (advisory_type_id, name, description, is_bullet_idea, status, published_at)
    VALUES (
        stock_basket_type_id,
        'Mid Cap Growth Picks',
        'High-growth mid-cap stocks with strong momentum',
        true,
        'published',
        CURRENT_TIMESTAMP
    );

    -- Stock Basket Items
    -- For Large Cap Opportunities
    INSERT INTO stock_basket_items (stock_basket_id, stock_name, stock_symbol, cmp, target, stop_loss, action, rationale, is_bullet_idea, display_order, status)
    VALUES 
        ((SELECT id FROM stock_baskets WHERE name = 'Large Cap Opportunities' LIMIT 1), 'Reliance Industries', 'RELIANCE', 2456.75, 2650.00, 2350.00, 'buy', 'Strong refining margins and retail expansion driving growth', false, 1, 'active'),
        ((SELECT id FROM stock_baskets WHERE name = 'Large Cap Opportunities' LIMIT 1), 'TCS', 'TCS', 3456.25, 3650.00, 3300.00, 'buy', 'Digital transformation deals and stable margins', false, 2, 'active'),
        ((SELECT id FROM stock_baskets WHERE name = 'Large Cap Opportunities' LIMIT 1), 'HDFC Bank', 'HDFCBANK', 1654.50, 1800.00, 1580.00, 'buy', 'Strong loan growth and improving asset quality', false, 3, 'active');

    -- For Mid Cap Growth (Bullet Ideas)
    INSERT INTO stock_basket_items (stock_basket_id, stock_name, stock_symbol, cmp, target, stop_loss, action, rationale, is_bullet_idea, display_order, status)
    VALUES 
        ((SELECT id FROM stock_baskets WHERE name = 'Mid Cap Growth Picks' AND is_bullet_idea = true LIMIT 1), 'Tata Motors', 'TATAMOTORS', 856.75, 950.00, 820.00, 'buy', 'EV transition and strong JLR performance', true, 1, 'active'),
        ((SELECT id FROM stock_baskets WHERE name = 'Mid Cap Growth Picks' AND is_bullet_idea = true LIMIT 1), 'Adani Ports', 'ADANIPORTS', 1234.50, 1350.00, 1180.00, 'buy', 'Port expansion and logistics growth', true, 2, 'active'),
        ((SELECT id FROM stock_baskets WHERE name = 'Mid Cap Growth Picks' AND is_bullet_idea = true LIMIT 1), 'Bajaj Finance', 'BAJFINANCE', 6789.25, 7200.00, 6500.00, 'buy', 'Strong AUM growth and digital lending', true, 3, 'active');

    -- ============================================
    -- 3. IPO ADVISORIES (Overview & Equities)
    -- ============================================
    INSERT INTO ipo_advisories (advisory_type_id, ipo_name, ipo_symbol, gmp, suggestion, lot_size, price_band_min, price_band_max, issue_date, issue_size, status, published_at)
    VALUES 
        (ipo_type_id, 'Motilal Oswal Business Cycle Fund', 'MOBCF', 125.50, 'Subscribe', 1, 10.00, 10.00, '25 Oct 2024 — 08 Nov 2024', '₹500 Cr', 'open', CURRENT_TIMESTAMP),
        (ipo_type_id, 'Tech Mahindra Digital', 'TMDIGITAL', 85.25, 'Subscribe', 1, 150.00, 155.00, '15 Nov 2024 — 20 Nov 2024', '₹800 Cr', 'upcoming', CURRENT_TIMESTAMP),
        (ipo_type_id, 'Green Energy Solutions', 'GREENENERGY', 45.75, 'Wait', 1, 200.00, 210.00, '01 Dec 2024 — 05 Dec 2024', '₹300 Cr', 'upcoming', CURRENT_TIMESTAMP);

    -- ============================================
    -- 4. ETF BASKETS (Equities)
    -- ============================================
    INSERT INTO etf_baskets (advisory_type_id, name, description, status, published_at)
    VALUES 
        (etf_basket_type_id, 'Nifty 50 ETF Basket', 'Top performing Nifty 50 ETFs', 'published', CURRENT_TIMESTAMP),
        (etf_basket_type_id, 'Sectoral ETF Picks', 'Best sectoral ETFs for current market cycle', 'published', CURRENT_TIMESTAMP);

    INSERT INTO etf_basket_items (etf_basket_id, name, symbol, cmp, target, action, rationale, display_order, status)
    VALUES 
        ((SELECT id FROM etf_baskets WHERE name = 'Nifty 50 ETF Basket' LIMIT 1), 'Nippon India Nifty 50 ETF', 'NIFTY50', 245.50, 260.00, 'buy', 'Low expense ratio and high liquidity', 1, 'active'),
        ((SELECT id FROM etf_baskets WHERE name = 'Nifty 50 ETF Basket' LIMIT 1), 'HDFC Nifty 50 ETF', 'HDFCNIFTY', 248.75, 265.00, 'buy', 'Strong tracking and consistent performance', 2, 'active'),
        ((SELECT id FROM etf_baskets WHERE name = 'Sectoral ETF Picks' LIMIT 1), 'Banking ETF', 'BANKETF', 156.25, 170.00, 'buy', 'Banking sector recovery and credit growth', 1, 'active');

    -- ============================================
    -- 5. MUTUAL FUND BASKETS (Overview & Mutual Funds)
    -- ============================================
    INSERT INTO mutual_fund_baskets (advisory_type_id, basket_type, name, description, max_schemes, status, published_at)
    VALUES 
        (mf_basket_type_id, 'general', 'Top Equity Funds', 'Best performing equity mutual funds', 10, 'published', CURRENT_TIMESTAMP),
        (mf_basket_type_id, 'sectoral', 'Banking Sector Funds', 'Top banking sector mutual funds', 5, 'published', CURRENT_TIMESTAMP);

    INSERT INTO mutual_fund_basket_items (mutual_fund_basket_id, scheme_name, scheme_code, entry_price, current_nav, trend, display_order, status)
    VALUES 
        ((SELECT id FROM mutual_fund_baskets WHERE name = 'Top Equity Funds' LIMIT 1), 'HDFC Top 100 Fund', 'HDFC100', 125.50, 132.75, 'positive', 1, 'active'),
        ((SELECT id FROM mutual_fund_baskets WHERE name = 'Top Equity Funds' LIMIT 1), 'SBI Bluechip Fund', 'SBIBLUECHIP', 98.25, 105.50, 'positive', 2, 'active'),
        ((SELECT id FROM mutual_fund_baskets WHERE name = 'Top Equity Funds' LIMIT 1), 'ICICI Prudential Value Discovery', 'ICICIVALUE', 145.75, 152.30, 'positive', 3, 'active'),
        ((SELECT id FROM mutual_fund_baskets WHERE name = 'Banking Sector Funds' LIMIT 1), 'SBI Banking & Financial Services Fund', 'SBIBANKING', 78.50, 82.25, 'positive', 1, 'active'),
        ((SELECT id FROM mutual_fund_baskets WHERE name = 'Banking Sector Funds' LIMIT 1), 'ICICI Prudential Banking & Financial Services', 'ICICIBANKING', 65.25, 68.90, 'positive', 2, 'active');

    -- ============================================
    -- 6. MF SCHEMES (Mutual Funds)
    -- ============================================
    INSERT INTO mf_schemes (name, scheme_code, amc, category, type, current_nav, entry_price, exit_price, trend, verdict, rationale, status)
    VALUES 
        ('HDFC Top 100 Fund', 'HDFC100', 'HDFC Mutual Fund', 'Large Cap', 'Equity', 132.75, 125.50, NULL, 'positive', 'Buy', 'Strong large-cap exposure with consistent returns', 'active'),
        ('SBI Bluechip Fund', 'SBIBLUECHIP', 'SBI Mutual Fund', 'Large Cap', 'Equity', 105.50, 98.25, NULL, 'positive', 'Buy', 'Well-diversified large-cap portfolio', 'active'),
        ('ICICI Prudential Value Discovery', 'ICICIVALUE', 'ICICI Prudential Mutual Fund', 'Value', 'Equity', 152.30, 145.75, NULL, 'positive', 'Buy', 'Value investing approach with strong track record', 'active'),
        ('Axis Midcap Fund', 'AXISMIDCAP', 'Axis Mutual Fund', 'Mid Cap', 'Equity', 89.25, 85.00, NULL, 'positive', 'Buy', 'Mid-cap growth opportunities', 'active'),
        ('Kotak Emerging Equity Fund', 'KOTAKEMERGING', 'Kotak Mahindra Mutual Fund', 'Small Cap', 'Equity', 156.50, 148.75, NULL, 'positive', 'Hold', 'Small-cap exposure with higher volatility', 'active');

    -- ============================================
    -- 7. NFOS (Mutual Funds)
    -- ============================================
    INSERT INTO nfos (name, amc, category, type, open_date, close_date, timeline, summary, minimum_investment, verdict, rationale, status)
    VALUES 
        ('Motilal Oswal Business Cycle Fund', 'Motilal Oswal Mutual Fund', 'Equity', 'Equity', '2024-10-25', '2024-11-08', '25 Oct 2024 — 08 Nov 2024', 
         '["Identifies companies poised to benefit from specific phases of the economic cycle.", "Suitable for long-term investors (5+ years) looking for tactical alpha.", "Key Focus: Domestic manufacturing, premium consumption, and financial credit growth."]',
         5000.00, 'Subscribe', 'Strong fund house with proven track record in thematic investing', 'open'),
        ('HDFC Manufacturing Fund', 'HDFC Mutual Fund', 'Equity', 'Equity', '2024-11-15', '2024-11-29', '15 Nov 2024 — 29 Nov 2024',
         '["Focus on manufacturing sector growth.", "PLI scheme beneficiaries.", "Export-oriented companies."]',
         5000.00, 'Subscribe', 'Manufacturing theme aligned with government initiatives', 'upcoming'),
        ('ICICI Prudential Infrastructure Fund', 'ICICI Prudential Mutual Fund', 'Equity', 'Equity', '2024-12-01', '2024-12-15', '01 Dec 2024 — 15 Dec 2024',
         '["Infrastructure development theme.", "Roads, ports, and power sectors.", "Long-term growth potential."]',
         5000.00, 'Wait', 'Infrastructure theme but wait for better entry', 'upcoming');

    -- ============================================
    -- 8. WEBINARS (Data Room)
    -- ============================================
    INSERT INTO webinars (title, description, status, date, time, duration, registration_url, recording_url)
    VALUES 
        ('The manufacturing renaissance in India', 'Deep dive into manufacturing sector growth and opportunities', 'Live', CURRENT_TIMESTAMP, '15:00', '60 minutes', 'https://example.com/register/manufacturing', NULL),
        ('Post-Election market dynamics', 'Understanding market trends after elections', 'Upcoming', CURRENT_TIMESTAMP + INTERVAL '7 days', '16:00', '45 minutes', 'https://example.com/register/election', NULL),
        ('Cracking the Small-cap puzzle', 'How to identify winning small-cap stocks', 'Recorded', CURRENT_TIMESTAMP - INTERVAL '10 days', '14:00', '45 mins', NULL, 'https://example.com/recordings/smallcap'),
        ('Banking Sector Outlook 2025', 'Credit growth, NIMs, and asset quality trends', 'Upcoming', CURRENT_TIMESTAMP + INTERVAL '14 days', '17:00', '60 minutes', 'https://example.com/register/banking', NULL),
        ('FII Flow Analysis', 'Understanding foreign institutional investor behavior', 'Recorded', CURRENT_TIMESTAMP - INTERVAL '5 days', '15:30', '50 mins', NULL, 'https://example.com/recordings/fii');

    -- ============================================
    -- 9. WEEKLY MARKET MOOD (Overview & Data Room)
    -- ============================================
    INSERT INTO weekly_market_mood (title, week, points, sentiment, sentiment_label)
    VALUES 
        ('Steady consolidation amidst global volatility', 'Week of 21 Oct — 27 Oct 2024',
         '["Nifty holds key support at 24,000 despite persistent FII selling.", "Earnings season kick-off: Banking sectors show resilient NIMs.", "Manufacturing PMI at 58.5 indicates strong industrial activity.", "RBI policy stance remains accommodative with focus on growth."]',
         0.72, 'Bullish'),
        ('Market resilience in face of global headwinds', 'Week of 14 Oct — 20 Oct 2024',
         '["Domestic flows offset FII outflows.", "Strong Q2 earnings across sectors.", "Oil prices stabilize below $85/bbl."]',
         0.68, 'Cautious Neutral'),
        ('Consolidation phase before next leg up', 'Week of 07 Oct — 13 Oct 2024',
         '["Markets taking a breather after strong rally.", "Sector rotation visible from IT to Banking.", "FII selling pressure continues but DII buying strong."]',
         0.65, 'Neutral');

    -- ============================================
    -- 10. WEEKLY AUDIO (Data Room)
    -- ============================================
    INSERT INTO weekly_audio (title, description, duration, week, audio_url, thumbnail_url, posted_at, status)
    VALUES 
        ('Understanding the FII sell-off cycle', 'Deep dive into why FIIs are selling and when they might return', '8:42', 'Week of 21 Oct — 27 Oct 2024', 
         'https://example.com/audio/fii-selloff.mp3', 'https://example.com/thumbnails/fii-selloff.jpg', CURRENT_TIMESTAMP - INTERVAL '2 days', 'published'),
        ('Banking sector earnings analysis', 'Q2 FY25 earnings review for major banks', '12:15', 'Week of 14 Oct — 20 Oct 2024',
         'https://example.com/audio/banking-earnings.mp3', 'https://example.com/thumbnails/banking-earnings.jpg', CURRENT_TIMESTAMP - INTERVAL '9 days', 'published'),
        ('Small-cap valuation concerns', 'Are small-caps overvalued? Expert analysis', '10:30', 'Week of 07 Oct — 13 Oct 2024',
         'https://example.com/audio/smallcap-valuation.mp3', 'https://example.com/thumbnails/smallcap-valuation.jpg', CURRENT_TIMESTAMP - INTERVAL '16 days', 'published');

END $$;
