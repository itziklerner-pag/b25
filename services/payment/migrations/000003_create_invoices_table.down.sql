-- Drop invoices table
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TABLE IF EXISTS invoices CASCADE;
