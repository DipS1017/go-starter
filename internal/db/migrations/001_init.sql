-- Enable uuid-ossp for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;
CREATE EXTENSION IF NOT EXISTS postgis_tiger_geocoder;


CREATE TYPE user_status AS ENUM ('ACTIVE', 'SUSPENDED', 'DEACTIVATED');

CREATE TYPE user_role AS ENUM ('USER', 'ADMIN', 'MODERATOR');


CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    username TEXT UNIQUE,
    email TEXT UNIQUE NOT NULL,
    phone_number TEXT,
    is_email_verified BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    password TEXT,
    password_changed_at TIMESTAMPTZ,
    google_id TEXT UNIQUE,
    apple_id TEXT UNIQUE,
    role user_role NOT NULL DEFAULT 'USER',
    user_type TEXT,
    status user_status NOT NULL DEFAULT 'ACTIVE',
    profile_image TEXT,
    current_plan UUID REFERENCES plans(id) ON DELETE CASCADE,
    stripe_customer_id TEXT,
    sub_end_date TIMESTAMPTZ,
    reset_token TEXT,
    reset_expires TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_profile (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    profile_banner_url TEXT,
    bio TEXT,
    thought TEXT,
    professional_summary TEXT,
    experience_level TEXT,
    profession TEXT,
    profession_details JSONB,
    gender TEXT,
    date_of_birth DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- USER_SESSIONS table
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

-- MEDIA table
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT,
    url TEXT,
    TYPE TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- USER MEDIA table
CREATE TABLE user_media (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    TYPE TEXT,
    name TEXT,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, media_id)
);

-- -- USER_ACTIVITY table
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    ACTION TEXT,
    entity_id UUID,
    entity_type TEXT NOT NULL,
    action_details JSONB,
    ip_address INET,
    timestamp TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_media_type ON user_media(TYPE);

CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING GIN (lower(name) gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING GIN (lower(email) gin_trgm_ops);

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_timestamp BEFORE
UPDATE
    ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp();

---- create above / drop below ----
-- Drop all the funtions
