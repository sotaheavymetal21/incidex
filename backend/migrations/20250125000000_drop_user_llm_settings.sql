-- +goose Up
-- マイグレーション: AI/LLM 機能の削除
-- 日付: 2025-01-25
-- 説明: AI 機能の削除に伴い user_llm_settings テーブルを削除します

DROP TABLE IF EXISTS user_llm_settings;

-- +goose Down
-- このマイグレーションは意図的にロールバック不可です
-- AI/LLM 機能はコードベースから削除されました
