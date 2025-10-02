-- User Profile Service Database Schema
-- PostgreSQL 14+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum for privacy levels
CREATE TYPE privacy_level AS ENUM ('public', 'friends', 'private');

-- User profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    preferences JSONB DEFAULT '{}'::jsonb,
    privacy_settings JSONB DEFAULT '{
        "profileVisibility": "public",
        "showEmail": false,
        "showBio": true,
        "showAvatar": true,
        "allowMessaging": true,
        "allowFollowing": true
    }'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT name_not_empty CHECK (LENGTH(TRIM(name)) > 0),
    CONSTRAINT bio_length CHECK (LENGTH(bio) <= 5000),
    CONSTRAINT avatar_url_format CHECK (avatar_url IS NULL OR avatar_url ~ '^https?://')
);

-- Indexes for performance
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_created_at ON user_profiles(created_at DESC);
CREATE INDEX idx_user_profiles_updated_at ON user_profiles(updated_at DESC);
CREATE INDEX idx_user_profiles_privacy ON user_profiles((privacy_settings->>'profileVisibility'));

-- GIN index for JSONB fields for efficient querying
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING GIN (preferences);
CREATE INDEX idx_user_profiles_privacy_settings ON user_profiles USING GIN (privacy_settings);

-- Full-text search index on name and bio
CREATE INDEX idx_user_profiles_search ON user_profiles USING GIN (
    to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(bio, ''))
);

-- Profile activity log table (optional, for audit trail)
CREATE TABLE IF NOT EXISTS profile_activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    changed_fields JSONB,
    changed_by VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_activity_log_profile_id ON profile_activity_log(profile_id);
CREATE INDEX idx_activity_log_created_at ON profile_activity_log(created_at DESC);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update updated_at on user_profiles
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to log profile changes
CREATE OR REPLACE FUNCTION log_profile_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO profile_activity_log (profile_id, action, changed_fields)
        VALUES (
            NEW.id,
            'update',
            jsonb_build_object(
                'old', row_to_json(OLD),
                'new', row_to_json(NEW)
            )
        );
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO profile_activity_log (profile_id, action, changed_fields)
        VALUES (NEW.id, 'create', row_to_json(NEW)::jsonb);
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO profile_activity_log (profile_id, action, changed_fields)
        VALUES (OLD.id, 'delete', row_to_json(OLD)::jsonb);
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to log profile activity (optional, can be disabled for performance)
-- Uncomment to enable activity logging
-- CREATE TRIGGER log_user_profile_changes
--     AFTER INSERT OR UPDATE OR DELETE ON user_profiles
--     FOR EACH ROW
--     EXECUTE FUNCTION log_profile_changes();

-- Create default admin/system user profile (optional)
-- INSERT INTO user_profiles (user_id, name, bio)
-- VALUES ('system', 'System', 'System administrator account')
-- ON CONFLICT (user_id) DO NOTHING;
