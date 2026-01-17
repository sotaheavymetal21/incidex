-- +goose Up
-- +goose StatementBegin

-- Drop template_tags table first (has foreign key to incident_templates)
DROP TABLE IF EXISTS template_tags;

-- Drop incident_templates table
DROP TABLE IF EXISTS incident_templates;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate incident_templates table
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

-- Recreate template_tags table
CREATE TABLE IF NOT EXISTS template_tags (
    incident_template_id INTEGER NOT NULL REFERENCES incident_templates(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (incident_template_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_template_tags_incident_template_id ON template_tags(incident_template_id);
CREATE INDEX IF NOT EXISTS idx_template_tags_tag_id ON template_tags(tag_id);

-- +goose StatementEnd
