-- +goose Up
-- Migration: Add constraints and improvements
-- Date: 2025-01-24
-- Description: Adds CHECK constraints, UNIQUE constraints, and creates user_llm_settings table

-- ============================================
-- 1. Add UNIQUE constraint to notification_settings.user_id
-- ============================================
-- notification_settings should be 1:1 with users
ALTER TABLE notification_settings
    ADD CONSTRAINT uq_notification_settings_user_id UNIQUE (user_id);

-- ============================================
-- 2. Change action_items.related_links from TEXT to JSONB
-- ============================================
-- First, convert existing data to valid JSON (empty array if null or empty)
UPDATE action_items
SET related_links = '[]'
WHERE related_links IS NULL OR related_links = '';

-- Alter the column type to JSONB
ALTER TABLE action_items
    ALTER COLUMN related_links TYPE JSONB USING related_links::jsonb;

-- Set default value for new rows
ALTER TABLE action_items
    ALTER COLUMN related_links SET DEFAULT '[]'::jsonb;

-- ============================================
-- 3. Create user_llm_settings table
-- ============================================
CREATE TABLE IF NOT EXISTS user_llm_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_type VARCHAR(50) NOT NULL,
    endpoint VARCHAR(500),
    model VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_llm_settings_user_id UNIQUE (user_id),
    CONSTRAINT chk_user_llm_settings_provider_type
        CHECK (provider_type IN ('openai', 'azure-openai', 'claude', 'ollama', 'custom'))
);

CREATE INDEX IF NOT EXISTS idx_user_llm_settings_user_id ON user_llm_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_llm_settings_provider_type ON user_llm_settings(provider_type);

-- ============================================
-- 4. Add CHECK constraints for enum-like columns
-- ============================================

-- users.role constraint
ALTER TABLE users
    ADD CONSTRAINT chk_users_role
    CHECK (role IN ('admin', 'editor', 'viewer'));

-- incidents.severity constraint
ALTER TABLE incidents
    ADD CONSTRAINT chk_incidents_severity
    CHECK (severity IN ('critical', 'high', 'medium', 'low'));

-- incidents.status constraint
ALTER TABLE incidents
    ADD CONSTRAINT chk_incidents_status
    CHECK (status IN ('open', 'investigating', 'resolved', 'closed'));

-- post_mortems.status constraint
ALTER TABLE post_mortems
    ADD CONSTRAINT chk_post_mortems_status
    CHECK (status IN ('draft', 'published'));

-- action_items.priority constraint
ALTER TABLE action_items
    ADD CONSTRAINT chk_action_items_priority
    CHECK (priority IN ('high', 'medium', 'low'));

-- action_items.status constraint
ALTER TABLE action_items
    ADD CONSTRAINT chk_action_items_status
    CHECK (status IN ('pending', 'in_progress', 'completed'));

-- ============================================
-- 5. Add comment for assignee_id clarification
-- ============================================
COMMENT ON COLUMN incidents.assignee_id IS 'Primary assignee for the incident. Use incident_assignees table for additional assignees.';

-- +goose Down
-- Rollback in reverse order

-- Remove comment
COMMENT ON COLUMN incidents.assignee_id IS NULL;

-- Remove CHECK constraints
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS chk_action_items_status;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS chk_action_items_priority;
ALTER TABLE post_mortems DROP CONSTRAINT IF EXISTS chk_post_mortems_status;
ALTER TABLE incidents DROP CONSTRAINT IF EXISTS chk_incidents_status;
ALTER TABLE incidents DROP CONSTRAINT IF EXISTS chk_incidents_severity;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

-- Drop user_llm_settings table
DROP TABLE IF EXISTS user_llm_settings;

-- Revert action_items.related_links to TEXT
ALTER TABLE action_items
    ALTER COLUMN related_links DROP DEFAULT;
ALTER TABLE action_items
    ALTER COLUMN related_links TYPE TEXT USING related_links::text;

-- Remove UNIQUE constraint from notification_settings
ALTER TABLE notification_settings
    DROP CONSTRAINT IF EXISTS uq_notification_settings_user_id;
