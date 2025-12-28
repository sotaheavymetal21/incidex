-- +goose Up
-- Migration: Add Full-Text Search Support
-- Date: 2025-01-01
-- Description: Adds tsvector column and GIN index for full-text search on incidents table

-- Enable pg_trgm extension for better Japanese/multilingual support
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add search_vector column to incidents table
ALTER TABLE incidents
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_incidents_search_vector
ON incidents USING GIN(search_vector);

-- Create function to update search_vector
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_incidents_search_vector()
RETURNS TRIGGER AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
    setweight(to_tsvector('simple', COALESCE(NEW.summary, '')), 'C') ||
    setweight(to_tsvector('simple', COALESCE(NEW.impact_scope, '')), 'D');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create trigger to automatically update search_vector on INSERT or UPDATE
DROP TRIGGER IF EXISTS incidents_search_vector_update ON incidents;
CREATE TRIGGER incidents_search_vector_update
  BEFORE INSERT OR UPDATE ON incidents
  FOR EACH ROW
  EXECUTE FUNCTION update_incidents_search_vector();

-- Update existing rows with search_vector
UPDATE incidents
SET search_vector =
  setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
  setweight(to_tsvector('simple', COALESCE(description, '')), 'B') ||
  setweight(to_tsvector('simple', COALESCE(summary, '')), 'C') ||
  setweight(to_tsvector('simple', COALESCE(impact_scope, '')), 'D');

-- Add comments for documentation
COMMENT ON COLUMN incidents.search_vector IS 'Full-text search vector combining title, description, summary, and impact_scope';
COMMENT ON INDEX idx_incidents_search_vector IS 'GIN index for full-text search performance';
COMMENT ON FUNCTION update_incidents_search_vector() IS 'Trigger function to automatically update search_vector on incidents';

-- +goose Down
-- Remove full-text search functionality
DROP TRIGGER IF EXISTS incidents_search_vector_update ON incidents;
DROP FUNCTION IF EXISTS update_incidents_search_vector();
DROP INDEX IF EXISTS idx_incidents_search_vector;
ALTER TABLE incidents DROP COLUMN IF EXISTS search_vector;
DROP EXTENSION IF EXISTS pg_trgm;
