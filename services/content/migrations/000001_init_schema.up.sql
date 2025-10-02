-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Users Table
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'author',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT users_role_check CHECK (role IN ('admin', 'editor', 'author', 'viewer'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active) WHERE is_active = true;

-- ============================================================================
-- Content Table
-- ============================================================================
CREATE TABLE content (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    body TEXT,
    excerpt TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tags TEXT[] DEFAULT '{}',
    categories TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    media_url TEXT,
    media_type VARCHAR(100),
    media_size BIGINT DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    view_count BIGINT NOT NULL DEFAULT 0,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT content_type_check CHECK (type IN ('post', 'article', 'media')),
    CONSTRAINT content_status_check CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT content_slug_unique UNIQUE (slug)
);

CREATE INDEX idx_content_type ON content(type);
CREATE INDEX idx_content_status ON content(status);
CREATE INDEX idx_content_author ON content(author_id);
CREATE INDEX idx_content_slug ON content(slug);
CREATE INDEX idx_content_tags ON content USING GIN(tags);
CREATE INDEX idx_content_categories ON content USING GIN(categories);
CREATE INDEX idx_content_created_at ON content(created_at DESC);
CREATE INDEX idx_content_updated_at ON content(updated_at DESC);
CREATE INDEX idx_content_published_at ON content(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_content_view_count ON content(view_count DESC);
CREATE INDEX idx_content_metadata ON content USING GIN(metadata);

-- Full-text search index
CREATE INDEX idx_content_search ON content USING GIN(
    to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(body, '') || ' ' || COALESCE(excerpt, ''))
);

-- ============================================================================
-- Content Versions Table
-- ============================================================================
CREATE TABLE content_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    excerpt TEXT,
    metadata JSONB DEFAULT '{}',
    changed_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    change_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT content_versions_unique UNIQUE (content_id, version)
);

CREATE INDEX idx_content_versions_content_id ON content_versions(content_id);
CREATE INDEX idx_content_versions_version ON content_versions(version);
CREATE INDEX idx_content_versions_created_at ON content_versions(created_at DESC);

-- ============================================================================
-- Triggers
-- ============================================================================

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to users table
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to content table
CREATE TRIGGER update_content_updated_at
    BEFORE UPDATE ON content
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Auto-increment version and save to version history
CREATE OR REPLACE FUNCTION save_content_version()
RETURNS TRIGGER AS $$
BEGIN
    -- If content is being updated (not created)
    IF TG_OP = 'UPDATE' AND (
        OLD.title IS DISTINCT FROM NEW.title OR
        OLD.body IS DISTINCT FROM NEW.body OR
        OLD.excerpt IS DISTINCT FROM NEW.excerpt OR
        OLD.metadata IS DISTINCT FROM NEW.metadata
    ) THEN
        -- Increment version
        NEW.version = OLD.version + 1;

        -- Save previous version to history
        INSERT INTO content_versions (
            content_id, version, title, body, excerpt, metadata, changed_by
        ) VALUES (
            OLD.id, OLD.version, OLD.title, OLD.body, OLD.excerpt, OLD.metadata, NEW.author_id
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER save_content_version_trigger
    BEFORE UPDATE ON content
    FOR EACH ROW
    EXECUTE FUNCTION save_content_version();

-- ============================================================================
-- Views
-- ============================================================================

-- Published content view
CREATE VIEW v_published_content AS
SELECT
    c.*,
    u.username as author_name,
    u.email as author_email
FROM content c
JOIN users u ON c.author_id = u.id
WHERE c.status = 'published'
  AND c.published_at IS NOT NULL
  AND c.published_at <= NOW();

-- Content with author details
CREATE VIEW v_content_with_author AS
SELECT
    c.*,
    u.username as author_name,
    u.email as author_email,
    u.role as author_role
FROM content c
JOIN users u ON c.author_id = u.id;

-- ============================================================================
-- Seed Data (Optional - for development)
-- ============================================================================

-- Create default admin user (password: admin123)
-- Password hash is bcrypt hash of 'admin123'
INSERT INTO users (email, username, password_hash, role) VALUES
('admin@example.com', 'admin', '$2a$10$rZ3L5YCqz5yBKNLGJfKGd.xQCqXVYE6O5F5SqB8qN5.O5F5SqB8qN', 'admin')
ON CONFLICT (email) DO NOTHING;
