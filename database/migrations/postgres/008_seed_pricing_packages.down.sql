-- Migration: Rollback Seed Pricing Packages
-- Description: Removes the seeded pricing packages

-- Delete all pricing packages that match the seeded data
DELETE FROM pricing_packages 
WHERE name IN ('Standard', 'Plus', 'Premium')
  AND duration_type IN ('quarterly', 'annual')
  AND status = 'active'
  AND is_published = true;

