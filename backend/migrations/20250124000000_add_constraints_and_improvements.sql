-- +goose Up
-- マイグレーション: 制約と改善の追加
-- 日付: 2025-01-24
-- 説明: CHECK 制約と UNIQUE 制約を追加します

-- ============================================
-- 1. notification_settings.user_id に UNIQUE 制約を追加
-- ============================================
-- notification_settings はユーザーと1対1の関係にする必要があります
ALTER TABLE notification_settings
    ADD CONSTRAINT uq_notification_settings_user_id UNIQUE (user_id);

-- ============================================
-- 2. action_items.related_links を TEXT から JSONB に変更
-- ============================================
-- まず、既存のデータを有効な JSON に変換します（null または空の場合は空配列に）
UPDATE action_items
SET related_links = '[]'
WHERE related_links IS NULL OR related_links = '';

-- カラムの型を JSONB に変更します
ALTER TABLE action_items
    ALTER COLUMN related_links TYPE JSONB USING related_links::jsonb;

-- 新しい行のデフォルト値を設定します
ALTER TABLE action_items
    ALTER COLUMN related_links SET DEFAULT '[]'::jsonb;

-- ============================================
-- 3. 列挙型カラムに CHECK 制約を追加
-- ============================================

-- users.role の制約
ALTER TABLE users
    ADD CONSTRAINT chk_users_role
    CHECK (role IN ('admin', 'editor', 'viewer'));

-- incidents.severity の制約
ALTER TABLE incidents
    ADD CONSTRAINT chk_incidents_severity
    CHECK (severity IN ('critical', 'high', 'medium', 'low'));

-- incidents.status の制約
ALTER TABLE incidents
    ADD CONSTRAINT chk_incidents_status
    CHECK (status IN ('open', 'investigating', 'resolved', 'closed'));

-- post_mortems.status の制約
ALTER TABLE post_mortems
    ADD CONSTRAINT chk_post_mortems_status
    CHECK (status IN ('draft', 'published'));

-- action_items.priority の制約
ALTER TABLE action_items
    ADD CONSTRAINT chk_action_items_priority
    CHECK (priority IN ('high', 'medium', 'low'));

-- action_items.status の制約
ALTER TABLE action_items
    ADD CONSTRAINT chk_action_items_status
    CHECK (status IN ('pending', 'in_progress', 'completed'));

-- ============================================
-- 4. assignee_id の説明コメントを追加
-- ============================================
COMMENT ON COLUMN incidents.assignee_id IS 'インシデントの主担当者。追加の担当者は incident_assignees テーブルを使用します。';

-- +goose Down
-- 逆順でロールバックします

-- コメントを削除
COMMENT ON COLUMN incidents.assignee_id IS NULL;

-- CHECK 制約を削除
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS chk_action_items_status;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS chk_action_items_priority;
ALTER TABLE post_mortems DROP CONSTRAINT IF EXISTS chk_post_mortems_status;
ALTER TABLE incidents DROP CONSTRAINT IF EXISTS chk_incidents_status;
ALTER TABLE incidents DROP CONSTRAINT IF EXISTS chk_incidents_severity;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

-- action_items.related_links を TEXT に戻す
ALTER TABLE action_items
    ALTER COLUMN related_links DROP DEFAULT;
ALTER TABLE action_items
    ALTER COLUMN related_links TYPE TEXT USING related_links::text;

-- notification_settings から UNIQUE 制約を削除
ALTER TABLE notification_settings
    DROP CONSTRAINT IF EXISTS uq_notification_settings_user_id;
