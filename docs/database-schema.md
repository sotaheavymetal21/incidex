# incidex テーブル定義書

## 1. 概要

### 1.1 データベース情報
- **DBMS**: PostgreSQL 15+
- **文字コード**: UTF-8
- **タイムゾーン**: UTC

### 1.2 命名規則
- **テーブル名**: スネークケース、複数形
- **カラム名**: スネークケース
- **主キー**: `id` (BIGSERIAL)
- **外部キー**: `{テーブル名}_id`
- **タイムスタンプ**: `created_at`, `updated_at` (TIMESTAMP WITH TIME ZONE)

---

## 2. テーブル定義

### 2.1 users（ユーザー）

ユーザー情報を管理するテーブル。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | ユーザーID |
| email | VARCHAR(255) | NOT NULL, UNIQUE | メールアドレス |
| password_hash | VARCHAR(255) | NOT NULL | パスワードハッシュ（bcrypt） |
| name | VARCHAR(255) | NOT NULL | ユーザー名 |
| role | VARCHAR(20) | NOT NULL, DEFAULT 'viewer' | 役割（admin/editor/viewer） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |
| updated_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 更新日時 |

**インデックス**:
- `idx_users_email`: `email` (UNIQUE)
- `idx_users_role`: `role`

**制約**:
- `role` は 'admin', 'editor', 'viewer' のいずれか

---

### 2.2 incidents（インシデント）

インシデント情報を管理するテーブル。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | インシデントID |
| title | VARCHAR(500) | NOT NULL | タイトル |
| description | TEXT | NOT NULL | 詳細説明 |
| summary | VARCHAR(300) | NULL | AI要約（300文字以内） |
| severity | VARCHAR(20) | NOT NULL | 深刻度（critical/high/medium/low） |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'open' | ステータス（open/investigating/resolved/closed） |
| impact_scope | VARCHAR(500) | NULL | 影響範囲 |
| detected_at | TIMESTAMP WITH TIME ZONE | NOT NULL | 検知日時 |
| resolved_at | TIMESTAMP WITH TIME ZONE | NULL | 解決日時 |
| assignee_id | BIGINT | NULL | 担当者ID（users.idへの外部キー） |
| creator_id | BIGINT | NOT NULL | 作成者ID（users.idへの外部キー） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |
| updated_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 更新日時 |

**インデックス**:
- `idx_incidents_severity`: `severity`
- `idx_incidents_status`: `status`
- `idx_incidents_detected_at`: `detected_at`
- `idx_incidents_resolved_at`: `resolved_at`
- `idx_incidents_assignee_id`: `assignee_id`
- `idx_incidents_creator_id`: `creator_id`
- `idx_incidents_created_at`: `created_at`
- `idx_incidents_title_description`: GIN全文検索インデックス（Phase 2）

**外部キー**:
- `fk_incidents_assignee`: `assignee_id` → `users.id` (ON DELETE SET NULL)
- `fk_incidents_creator`: `creator_id` → `users.id` (ON DELETE RESTRICT)

**制約**:
- `severity` は 'critical', 'high', 'medium', 'low' のいずれか
- `status` は 'open', 'investigating', 'resolved', 'closed' のいずれか
- `resolved_at` は `detected_at` 以降であること

---

### 2.3 tags（タグ）

タグ情報を管理するテーブル。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | タグID |
| name | VARCHAR(100) | NOT NULL, UNIQUE | タグ名 |
| color | VARCHAR(7) | NOT NULL, DEFAULT '#000000' | カラーコード（Hex形式） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |
| updated_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 更新日時 |

**インデックス**:
- `idx_tags_name`: `name` (UNIQUE)

**制約**:
- `color` は Hex形式（#RRGGBB）であること

---

### 2.4 incident_tags（インシデント-タグ関連）

インシデントとタグの多対多関係を管理するテーブル。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| incident_id | BIGINT | NOT NULL | インシデントID（incidents.idへの外部キー） |
| tag_id | BIGINT | NOT NULL | タグID（tags.idへの外部キー） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |

**主キー**: `(incident_id, tag_id)`

**インデックス**:
- `idx_incident_tags_incident_id`: `incident_id`
- `idx_incident_tags_tag_id`: `tag_id`

**外部キー**:
- `fk_incident_tags_incident`: `incident_id` → `incidents.id` (ON DELETE CASCADE)
- `fk_incident_tags_tag`: `tag_id` → `tags.id` (ON DELETE CASCADE)

---

### 2.5 timeline_events（タイムラインイベント）

インシデントのタイムラインイベントを管理するテーブル。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | イベントID |
| incident_id | BIGINT | NOT NULL | インシデントID（incidents.idへの外部キー） |
| event_type | VARCHAR(50) | NOT NULL | イベントタイプ |
| event_time | TIMESTAMP WITH TIME ZONE | NOT NULL | イベント時刻 |
| description | TEXT | NOT NULL | 説明 |
| recorder_id | BIGINT | NOT NULL | 記録者ID（users.idへの外部キー） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |
| updated_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 更新日時 |

**インデックス**:
- `idx_timeline_events_incident_id`: `incident_id`
- `idx_timeline_events_event_time`: `event_time`
- `idx_timeline_events_recorder_id`: `recorder_id`

**外部キー**:
- `fk_timeline_events_incident`: `incident_id` → `incidents.id` (ON DELETE CASCADE)
- `fk_timeline_events_recorder`: `recorder_id` → `users.id` (ON DELETE RESTRICT)

**制約**:
- `event_type` は 'detected', 'investigation_started', 'root_cause_identified', 'mitigation', 'resolved', 'other' のいずれか

---

### 2.6 postmortems（ポストモーテム）

ポストモーテム情報を管理するテーブル（Phase 2）。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | ポストモーテムID |
| incident_id | BIGINT | NOT NULL, UNIQUE | インシデントID（incidents.idへの外部キー） |
| root_cause | TEXT | NULL | 根本原因 |
| good_points | TEXT | NULL | 対応の良かった点 |
| improvement_points | TEXT | NULL | 改善が必要だった点 |
| prevention_measures | TEXT | NULL | 再発防止策 |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |
| updated_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 更新日時 |

**インデックス**:
- `idx_postmortems_incident_id`: `incident_id` (UNIQUE)

**外部キー**:
- `fk_postmortems_incident`: `incident_id` → `incidents.id` (ON DELETE CASCADE)

**制約**:
- `incident_id` は一意（1インシデントに1ポストモーテム）

---

### 2.7 action_items（アクションアイテム）

ポストモーテムのアクションアイテムを管理するテーブル（Phase 2）。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | アクションアイテムID |
| postmortem_id | BIGINT | NOT NULL | ポストモーテムID（postmortems.idへの外部キー） |
| description | TEXT | NOT NULL | 説明 |
| assignee_id | BIGINT | NULL | 担当者ID（users.idへの外部キー） |
| due_date | TIMESTAMP WITH TIME ZONE | NULL | 期限 |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | ステータス（pending/in_progress/completed/cancelled） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |
| updated_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 更新日時 |

**インデックス**:
- `idx_action_items_postmortem_id`: `postmortem_id`
- `idx_action_items_assignee_id`: `assignee_id`
- `idx_action_items_status`: `status`
- `idx_action_items_due_date`: `due_date`

**外部キー**:
- `fk_action_items_postmortem`: `postmortem_id` → `postmortems.id` (ON DELETE CASCADE)
- `fk_action_items_assignee`: `assignee_id` → `users.id` (ON DELETE SET NULL)

**制約**:
- `status` は 'pending', 'in_progress', 'completed', 'cancelled' のいずれか

---

### 2.8 files（ファイル）

インシデントに添付されたファイル情報を管理するテーブル（Phase 3）。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | ファイルID |
| incident_id | BIGINT | NOT NULL | インシデントID（incidents.idへの外部キー） |
| file_name | VARCHAR(500) | NOT NULL | ファイル名 |
| file_path | VARCHAR(1000) | NOT NULL | ファイルパス（MinIO内） |
| file_size | BIGINT | NOT NULL | ファイルサイズ（バイト） |
| mime_type | VARCHAR(100) | NOT NULL | MIMEタイプ |
| uploaded_by | BIGINT | NOT NULL | アップロード者ID（users.idへの外部キー） |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |

**インデックス**:
- `idx_files_incident_id`: `incident_id`
- `idx_files_uploaded_by`: `uploaded_by`

**外部キー**:
- `fk_files_incident`: `incident_id` → `incidents.id` (ON DELETE CASCADE)
- `fk_files_uploaded_by`: `uploaded_by` → `users.id` (ON DELETE RESTRICT)

**制約**:
- `file_size` は 50MB（52,428,800バイト）以下であること
- 1インシデントあたりの合計ファイルサイズは500MB（524,288,000バイト）以下であること

---

### 2.9 search_history（検索履歴）

検索履歴を管理するテーブル（Phase 2、オプション）。

| カラム名 | データ型 | 制約 | 説明 |
|---------|---------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 検索履歴ID |
| user_id | BIGINT | NOT NULL | ユーザーID（users.idへの外部キー） |
| search_query | TEXT | NOT NULL | 検索クエリ |
| filters | JSONB | NULL | フィルタ条件（JSON形式） |
| result_count | INTEGER | NOT NULL | 検索結果件数 |
| created_at | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT NOW() | 作成日時 |

**インデックス**:
- `idx_search_history_user_id`: `user_id`
- `idx_search_history_created_at`: `created_at`

**外部キー**:
- `fk_search_history_user`: `user_id` → `users.id` (ON DELETE CASCADE)

---

## 3. ビュー定義

### 3.1 incident_summary_view（インシデントサマリービュー）

インシデント一覧表示用のビュー。

```sql
CREATE VIEW incident_summary_view AS
SELECT 
    i.id,
    i.title,
    i.summary,
    i.severity,
    i.status,
    i.impact_scope,
    i.detected_at,
    i.resolved_at,
    i.created_at,
    i.updated_at,
    a.id AS assignee_id,
    a.name AS assignee_name,
    a.email AS assignee_email,
    c.id AS creator_id,
    c.name AS creator_name,
    c.email AS creator_email
FROM incidents i
LEFT JOIN users a ON i.assignee_id = a.id
LEFT JOIN users c ON i.creator_id = c.id;
```

---

## 4. トリガー定義

### 4.1 update_timestamp_trigger

`updated_at` カラムを自動更新するトリガー。

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 各テーブルにトリガーを設定
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_incidents_updated_at BEFORE UPDATE ON incidents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_timeline_events_updated_at BEFORE UPDATE ON timeline_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_postmortems_updated_at BEFORE UPDATE ON postmortems
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_action_items_updated_at BEFORE UPDATE ON action_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

## 5. 全文検索設定（Phase 2）

### 5.1 全文検索用カラム追加

```sql
-- incidentsテーブルに全文検索用カラムを追加
ALTER TABLE incidents ADD COLUMN search_vector tsvector;

-- 全文検索インデックス作成
CREATE INDEX idx_incidents_search_vector ON incidents USING GIN(search_vector);

-- 全文検索ベクトル更新関数
CREATE OR REPLACE FUNCTION incidents_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.summary, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- トリガー設定
CREATE TRIGGER incidents_search_vector_update_trigger
    BEFORE INSERT OR UPDATE ON incidents
    FOR EACH ROW EXECUTE FUNCTION incidents_search_vector_update();
```

---

## 6. データ容量見積もり

### 6.1 テーブル別容量見積もり

| テーブル | 1レコード平均サイズ | 年間想定件数 | 年間容量 |
|---------|-------------------|------------|---------|
| users | 500 bytes | 50件 | 25 KB |
| incidents | 10 KB | 10,000件 | 100 MB |
| tags | 200 bytes | 50件 | 10 KB |
| incident_tags | 50 bytes | 30,000件（1インシデントあたり3タグ） | 1.5 MB |
| timeline_events | 2 KB | 50,000件（1インシデントあたり5イベント） | 100 MB |
| postmortems | 5 KB | 5,000件（解決済みの50%） | 25 MB |
| action_items | 1 KB | 10,000件（1ポストモーテムあたり2アイテム） | 10 MB |
| files | 500 bytes（メタデータのみ） | 10,000件 | 5 MB |
| search_history | 1 KB | 50,000件 | 50 MB |

**合計**: 約300 MB（ファイルストレージ除く）

---

## 7. バックアップ・復旧

### 7.1 バックアップ方針
- 日次フルバックアップ
- 1時間ごとのWAL（Write-Ahead Logging）アーカイブ
- バックアップ保持期間: 30日間

### 7.2 復旧目標
- RPO（Recovery Point Objective）: 1時間
- RTO（Recovery Time Objective）: 4時間

