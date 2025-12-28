-- +goose Up
-- Migration: Create Initial Tables
-- Date: 2025-01-01
-- Description: Creates all initial tables for the Incidex application

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    employee_number VARCHAR(255) UNIQUE,
    department VARCHAR(255),
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_employee_number ON users(employee_number);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    color VARCHAR(7) DEFAULT '#808080',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- Incidents table
CREATE TABLE IF NOT EXISTS incidents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    summary VARCHAR(300),
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    impact_scope VARCHAR(500),
    detected_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    creator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    sla_target_resolution_hours INTEGER DEFAULT 0,
    sla_deadline TIMESTAMP,
    sla_violated BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incidents_title ON incidents(title);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_detected_at ON incidents(detected_at);
CREATE INDEX IF NOT EXISTS idx_incidents_assignee_id ON incidents(assignee_id);
CREATE INDEX IF NOT EXISTS idx_incidents_creator_id ON incidents(creator_id);
CREATE INDEX IF NOT EXISTS idx_incidents_created_at ON incidents(created_at);
CREATE INDEX IF NOT EXISTS idx_incidents_sla_deadline ON incidents(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_incidents_sla_violated ON incidents(sla_violated);

-- Incident Tags (many-to-many)
CREATE TABLE IF NOT EXISTS incident_tags (
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (incident_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_incident_tags_incident_id ON incident_tags(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_tags_tag_id ON incident_tags(tag_id);

-- Incident Assignees (many-to-many)
CREATE TABLE IF NOT EXISTS incident_assignees (
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (incident_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_incident_assignees_incident_id ON incident_assignees(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_assignees_user_id ON incident_assignees(user_id);

-- Incident Activities table
CREATE TABLE IF NOT EXISTS incident_activities (
    id SERIAL PRIMARY KEY,
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    activity_type VARCHAR(50) NOT NULL,
    comment TEXT,
    old_value VARCHAR(100),
    new_value VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incident_activities_incident_id ON incident_activities(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_activities_user_id ON incident_activities(user_id);
CREATE INDEX IF NOT EXISTS idx_incident_activities_activity_type ON incident_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_incident_activities_created_at ON incident_activities(created_at);

-- Attachments table
CREATE TABLE IF NOT EXISTS attachments (
    id SERIAL PRIMARY KEY,
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    storage_key VARCHAR(500) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attachments_incident_id ON attachments(incident_id);
CREATE INDEX IF NOT EXISTS idx_attachments_storage_key ON attachments(storage_key);

-- Notification Settings table
CREATE TABLE IF NOT EXISTS notification_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    slack_enabled BOOLEAN DEFAULT false,
    slack_webhook VARCHAR(512),
    notify_on_incident_created BOOLEAN DEFAULT true,
    notify_on_assigned BOOLEAN DEFAULT true,
    notify_on_comment BOOLEAN DEFAULT true,
    notify_on_status_change BOOLEAN DEFAULT true,
    notify_on_severity_change BOOLEAN DEFAULT true,
    notify_on_resolved BOOLEAN DEFAULT true,
    notify_on_escalation BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notification_settings_user_id ON notification_settings(user_id);

-- Incident Templates table
CREATE TABLE IF NOT EXISTS incident_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL,
    impact_scope VARCHAR(500),
    creator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_public BOOLEAN DEFAULT false,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incident_templates_name ON incident_templates(name);
CREATE INDEX IF NOT EXISTS idx_incident_templates_creator_id ON incident_templates(creator_id);
CREATE INDEX IF NOT EXISTS idx_incident_templates_is_public ON incident_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_incident_templates_created_at ON incident_templates(created_at);

-- Template Tags (many-to-many)
CREATE TABLE IF NOT EXISTS template_tags (
    incident_template_id INTEGER NOT NULL REFERENCES incident_templates(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (incident_template_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_template_tags_incident_template_id ON template_tags(incident_template_id);
CREATE INDEX IF NOT EXISTS idx_template_tags_tag_id ON template_tags(tag_id);

-- Post-Mortems table
CREATE TABLE IF NOT EXISTS post_mortems (
    id SERIAL PRIMARY KEY,
    incident_id INTEGER UNIQUE NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    root_cause TEXT,
    impact_analysis TEXT,
    what_went_well TEXT,
    what_went_wrong TEXT,
    lessons_learned TEXT,
    five_whys_analysis JSON,
    ai_root_cause_suggestion TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_post_mortems_incident_id ON post_mortems(incident_id);
CREATE INDEX IF NOT EXISTS idx_post_mortems_author_id ON post_mortems(author_id);
CREATE INDEX IF NOT EXISTS idx_post_mortems_status ON post_mortems(status);
CREATE INDEX IF NOT EXISTS idx_post_mortems_created_at ON post_mortems(created_at);

-- Action Items table
CREATE TABLE IF NOT EXISTS action_items (
    id SERIAL PRIMARY KEY,
    post_mortem_id INTEGER NOT NULL REFERENCES post_mortems(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    due_date TIMESTAMP,
    related_links TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_action_items_post_mortem_id ON action_items(post_mortem_id);
CREATE INDEX IF NOT EXISTS idx_action_items_assignee_id ON action_items(assignee_id);
CREATE INDEX IF NOT EXISTS idx_action_items_priority ON action_items(priority);
CREATE INDEX IF NOT EXISTS idx_action_items_status ON action_items(status);
CREATE INDEX IF NOT EXISTS idx_action_items_created_at ON action_items(created_at);

-- Audit Logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    user_name VARCHAR(255),
    user_email VARCHAR(255),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100),
    resource_id INTEGER,
    method VARCHAR(10),
    path VARCHAR(500),
    ip_address VARCHAR(45),
    user_agent TEXT,
    status_code INTEGER,
    details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- +goose Down
-- Drop tables in reverse order to respect foreign key constraints
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS action_items;
DROP TABLE IF EXISTS post_mortems;
DROP TABLE IF EXISTS template_tags;
DROP TABLE IF EXISTS incident_templates;
DROP TABLE IF EXISTS notification_settings;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS incident_activities;
DROP TABLE IF EXISTS incident_assignees;
DROP TABLE IF EXISTS incident_tags;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
