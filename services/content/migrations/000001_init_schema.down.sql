-- Drop views
DROP VIEW IF EXISTS v_content_with_author;
DROP VIEW IF EXISTS v_published_content;

-- Drop triggers
DROP TRIGGER IF EXISTS save_content_version_trigger ON content;
DROP TRIGGER IF EXISTS update_content_updated_at ON content;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop functions
DROP FUNCTION IF EXISTS save_content_version();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS content_versions;
DROP TABLE IF EXISTS content;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
