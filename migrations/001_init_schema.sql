-- WeChat Service Database Schema
-- PostgreSQL

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- USERS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    openid          VARCHAR(64) NOT NULL UNIQUE,
    unionid         VARCHAR(64),
    nickname        VARCHAR(256),
    sex             SMALLINT DEFAULT 0,
    city            VARCHAR(64),
    province        VARCHAR(64),
    country         VARCHAR(64),
    language        VARCHAR(32),
    head_img_url    VARCHAR(512),
    remark          VARCHAR(256),
    group_id        INTEGER DEFAULT 0,
    subscribe_time  TIMESTAMP WITH TIME ZONE,
    unsubscribe_time TIMESTAMP WITH TIME ZONE,
    subscribe_status SMALLINT DEFAULT 0,  -- 0: unsubscribed, 1: subscribed
    latitude        DECIMAL(10, 7),
    longitude       DECIMAL(10, 7),
    precision       DECIMAL(10, 5),
    tags            JSONB DEFAULT '[]',
    raw_data        JSONB DEFAULT '{}',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for users
CREATE INDEX IF NOT EXISTS idx_users_openid ON users(openid);
CREATE INDEX IF NOT EXISTS idx_users_subscribe_status ON users(subscribe_status);
CREATE INDEX IF NOT EXISTS idx_users_subscribe_time ON users(subscribe_time DESC);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Full text search index for nickname
CREATE INDEX IF NOT EXISTS idx_users_nickname_gin ON users USING GIN (to_tsvector('simple', COALESCE(nickname, '')));

-- =====================================================
-- MESSAGES TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS messages (
    id              BIGSERIAL PRIMARY KEY,
    msg_id          BIGINT NOT NULL,
    msg_data_id     VARCHAR(64),
    idx             SMALLINT DEFAULT 1,
    from_user       VARCHAR(64) NOT NULL,
    to_user         VARCHAR(64) NOT NULL,
    direction       VARCHAR(4) NOT NULL DEFAULT 'in',  -- 'in' or 'out'
    msg_type        VARCHAR(32) NOT NULL,
    content         TEXT,
    media_id        VARCHAR(256),
    thumb_media_id  VARCHAR(256),
    format          VARCHAR(16),
    pic_url         VARCHAR(512),
    location_x      DECIMAL(10, 7),
    location_y      DECIMAL(10, 7),
    scale           INTEGER,
    label           VARCHAR(256),
    title           VARCHAR(256),
    description     TEXT,
    url             VARCHAR(512),
    event_type      VARCHAR(32),
    event_key       VARCHAR(256),
    ticket          VARCHAR(128),
    menu_id         BIGINT,
    raw_data        JSONB DEFAULT '{}',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for messages
CREATE INDEX IF NOT EXISTS idx_messages_msg_id ON messages(msg_id);
CREATE INDEX IF NOT EXISTS idx_messages_from_user ON messages(from_user);
CREATE INDEX IF NOT EXISTS idx_messages_to_user ON messages(to_user);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_direction ON messages(direction);
CREATE INDEX IF NOT EXISTS idx_messages_msg_type ON messages(msg_type);

-- Composite index for user message listing
CREATE INDEX IF NOT EXISTS idx_messages_user_time ON messages(from_user, created_at DESC);

-- =====================================================
-- EVENTS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS events (
    id              BIGSERIAL PRIMARY KEY,
    openid          VARCHAR(64) NOT NULL,
    event_type      VARCHAR(32) NOT NULL,
    event_key       VARCHAR(256),
    ticket          VARCHAR(128),
    menu_id         BIGINT,
    latitude        DECIMAL(10, 7),
    longitude       DECIMAL(10, 7),
    precision       DECIMAL(10, 5),
    scan_type       VARCHAR(32),
    scan_result     TEXT,
    pic_count       SMALLINT DEFAULT 0,
    pic_md5_list    TEXT[],
    location_x      DECIMAL(10, 7),
    location_y      DECIMAL(10, 7),
    scale           INTEGER,
    address         VARCHAR(256),
    poi_name        VARCHAR(128),
    raw_data        JSONB DEFAULT '{}',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for events
CREATE INDEX IF NOT EXISTS idx_events_openid ON events(openid);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);

-- Composite index for event queries
CREATE INDEX IF NOT EXISTS idx_events_openid_time ON events(openid, created_at DESC);

-- =====================================================
-- MENUS TABLE (for conditional menu rules)
-- =====================================================
CREATE TABLE IF NOT EXISTS menu_rules (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(64) NOT NULL,
    menu_id         BIGINT NOT NULL,
    tag_id          INTEGER,
    sex             SMALLINT,
    client_platform VARCHAR(32),
    language        VARCHAR(16),
    country         VARCHAR(64),
    province        VARCHAR(64),
    city            VARCHAR(64),
    is_enabled      BOOLEAN DEFAULT TRUE,
    priority        INTEGER DEFAULT 0,
    match_count     BIGINT DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_menu_rules_is_enabled ON menu_rules(is_enabled);
CREATE INDEX IF NOT EXISTS idx_menu_rules_priority ON menu_rules(priority DESC);

-- =====================================================
-- STATISTICS TABLES
-- =====================================================
CREATE TABLE IF NOT EXISTS daily_stats (
    id              BIGSERIAL PRIMARY KEY,
    stat_date       DATE NOT NULL UNIQUE,
    new_subscribers BIGINT DEFAULT 0,
    unsubscribes    BIGINT DEFAULT 0,
    messages_in     BIGINT DEFAULT 0,
    messages_out    BIGINT DEFAULT 0,
    active_users    BIGINT DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(stat_date DESC);

-- =====================================================
-- FUNCTIONS
-- =====================================================

-- Update updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name FROM information_schema.columns
        WHERE column_name = 'updated_at'
        AND table_schema = 'public'
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS update_%s_updated_at ON %I', t, t);
        EXECUTE format('CREATE TRIGGER update_%s_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END;
$$;

-- Function to get user message count
CREATE OR REPLACE FUNCTION count_user_messages(p_openid VARCHAR)
RETURNS BIGINT AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM messages WHERE from_user = p_openid);
END;
$$ LANGUAGE plpgsql;

-- Function to increment daily stats
CREATE OR REPLACE FUNCTION increment_daily_stats(p_stat_date DATE, p_field VARCHAR)
RETURNS VOID AS $$
BEGIN
    INSERT INTO daily_stats (stat_date, new_subscribers, unsubscribes, messages_in, messages_out, active_users)
    VALUES (p_stat_date, 0, 0, 0, 0, 0)
    ON CONFLICT (stat_date) DO UPDATE SET
        new_subscribers = daily_stats.new_subscribers + (p_field = 'new_subscribers'::VARCHAR)::INT,
        unsubscribes = daily_stats.unsubscribes + (p_field = 'unsubscribes'::VARCHAR)::INT,
        messages_in = daily_stats.messages_in + (p_field = 'messages_in'::VARCHAR)::INT,
        messages_out = daily_stats.messages_out + (p_field = 'messages_out'::VARCHAR)::INT;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE users IS 'WeChat users who have interacted with the service account';
COMMENT ON TABLE messages IS 'All messages exchanged between users and the service account';
COMMENT ON TABLE events IS 'Events triggered by user actions (subscribe, scan, click, etc.)';
COMMENT ON TABLE menu_rules IS 'Conditional menu matching rules';
COMMENT ON TABLE daily_stats IS 'Daily statistics for analytics';
