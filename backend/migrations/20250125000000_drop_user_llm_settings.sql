-- +goose Up
-- Migration: Remove AI/LLM feature
-- Date: 2025-01-25
-- Description: Drops user_llm_settings table as AI feature is being removed

DROP TABLE IF EXISTS user_llm_settings;

-- +goose Down
-- This migration is intentionally not reversible
-- The AI/LLM feature has been removed from the codebase
