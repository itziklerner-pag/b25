-- Drop payment_methods table
DROP TRIGGER IF EXISTS update_payment_methods_updated_at ON payment_methods;
DROP TABLE IF EXISTS payment_methods CASCADE;
