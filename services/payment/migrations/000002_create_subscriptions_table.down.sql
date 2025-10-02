-- Drop subscriptions table
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TABLE IF EXISTS subscriptions CASCADE;
