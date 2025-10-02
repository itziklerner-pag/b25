-- Drop refunds table
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;
DROP TABLE IF EXISTS refunds CASCADE;
