-- +goose Up
-- マイグレーション: 全文検索サポートの追加
-- 日付: 2025-01-01
-- 説明: インシデントテーブルに全文検索用の tsvector カラムと GIN インデックスを追加します

-- 日本語/多言語対応を向上させるために pg_trgm 拡張を有効化します
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- インシデントテーブルに search_vector カラムを追加します
ALTER TABLE incidents
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 全文検索用の GIN インデックスを作成します
CREATE INDEX IF NOT EXISTS idx_incidents_search_vector
ON incidents USING GIN(search_vector);

-- search_vector を更新する関数を作成します
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_incidents_search_vector()
RETURNS TRIGGER AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
    setweight(to_tsvector('simple', COALESCE(NEW.impact_scope, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- INSERT または UPDATE 時に search_vector を自動更新するトリガーを作成します
DROP TRIGGER IF EXISTS incidents_search_vector_update ON incidents;
CREATE TRIGGER incidents_search_vector_update
  BEFORE INSERT OR UPDATE ON incidents
  FOR EACH ROW
  EXECUTE FUNCTION update_incidents_search_vector();

-- 既存の行の search_vector を更新します
UPDATE incidents
SET search_vector =
  setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
  setweight(to_tsvector('simple', COALESCE(description, '')), 'B') ||
  setweight(to_tsvector('simple', COALESCE(impact_scope, '')), 'C');

-- ドキュメント用のコメントを追加します
COMMENT ON COLUMN incidents.search_vector IS 'title、description、impact_scope を結合した全文検索用ベクトル';
COMMENT ON INDEX idx_incidents_search_vector IS '全文検索パフォーマンス向上用の GIN インデックス';
COMMENT ON FUNCTION update_incidents_search_vector() IS 'インシデントの search_vector を自動更新するトリガー関数';

-- +goose Down
-- 全文検索機能を削除します
DROP TRIGGER IF EXISTS incidents_search_vector_update ON incidents;
DROP FUNCTION IF EXISTS update_incidents_search_vector();
DROP INDEX IF EXISTS idx_incidents_search_vector;
ALTER TABLE incidents DROP COLUMN IF EXISTS search_vector;
DROP EXTENSION IF EXISTS pg_trgm;
