# incidex API仕様書

## 1. 概要

### 1.1 API基本情報
- **形式**: REST API
- **データ形式**: JSON
- **認証方式**: Bearer Token (JWT)
- **ベースURL**: `/api`

### 1.2 認証
すべてのAPIリクエスト（認証エンドポイントを除く）には、AuthorizationヘッダーにJWTトークンを含める必要があります。

```
Authorization: Bearer <JWT_TOKEN>
```

### 1.3 エラーレスポンス
エラー時は以下の形式で返却されます：

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ",
    "details": {}
  }
}
```

### 1.4 HTTPステータスコード
- `200 OK`: 成功
- `201 Created`: リソース作成成功
- `400 Bad Request`: リクエスト不正
- `401 Unauthorized`: 認証失敗
- `403 Forbidden`: 権限不足
- `404 Not Found`: リソース不存在
- `500 Internal Server Error`: サーバーエラー

---

## 2. 認証API

### 2.1 ユーザー登録
**エンドポイント**: `POST /api/auth/register`

**リクエスト**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "User Name"
}
```

**レスポンス** (201 Created):
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "role": "viewer",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 2.2 ログイン
**エンドポイント**: `POST /api/auth/login`

**リクエスト**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**レスポンス** (200 OK):
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "role": "editor"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-02T00:00:00Z"
}
```

### 2.3 トークン検証
**エンドポイント**: `GET /api/auth/verify`

**レスポンス** (200 OK):
```json
{
  "valid": true,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "role": "editor"
  }
}
```

---

## 3. インシデントAPI

### 3.1 インシデント一覧取得
**エンドポイント**: `GET /api/incidents`

**クエリパラメータ**:
- `page` (integer, default: 1): ページ番号
- `limit` (integer, default: 20): 1ページあたりの件数
- `severity` (string): 深刻度フィルタ (critical/high/medium/low)
- `status` (string): ステータスフィルタ (open/investigating/resolved/closed)
- `tag_ids` (string, comma-separated): タグIDのリスト
- `search` (string): タイトル・説明での部分一致検索
- `sort` (string, default: created_at): ソート項目 (created_at/detected_at/resolved_at)
- `order` (string, default: desc): ソート順 (asc/desc)

**レスポンス** (200 OK):
```json
{
  "incidents": [
    {
      "id": 1,
      "title": "Database connection timeout",
      "summary": "Database connection timeout occurred...",
      "severity": "high",
      "status": "resolved",
      "impact_scope": "Production API",
      "detected_at": "2024-01-01T10:00:00Z",
      "resolved_at": "2024-01-01T11:30:00Z",
      "assignee": {
        "id": 2,
        "name": "John Doe",
        "email": "john@example.com"
      },
      "creator": {
        "id": 1,
        "name": "Jane Smith",
        "email": "jane@example.com"
      },
      "tags": [
        {
          "id": 1,
          "name": "Database",
          "color": "#FF5733"
        }
      ],
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T11:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### 3.2 インシデント詳細取得
**エンドポイント**: `GET /api/incidents/:id`

**レスポンス** (200 OK):
```json
{
  "id": 1,
  "title": "Database connection timeout",
  "description": "詳細な説明...",
  "summary": "Database connection timeout occurred...",
  "severity": "high",
  "status": "resolved",
  "impact_scope": "Production API",
  "detected_at": "2024-01-01T10:00:00Z",
  "resolved_at": "2024-01-01T11:30:00Z",
  "assignee": {
    "id": 2,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "creator": {
    "id": 1,
    "name": "Jane Smith",
    "email": "jane@example.com"
  },
  "tags": [
    {
      "id": 1,
      "name": "Database",
      "color": "#FF5733"
    }
  ],
  "timeline_events": [
    {
      "id": 1,
      "event_type": "detected",
      "event_time": "2024-01-01T10:00:00Z",
      "description": "Issue detected by monitoring system",
      "recorder": {
        "id": 1,
        "name": "Jane Smith"
      }
    }
  ],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T11:30:00Z"
}
```

### 3.3 インシデント作成
**エンドポイント**: `POST /api/incidents`

**権限**: 編集者以上

**リクエスト**:
```json
{
  "title": "Database connection timeout",
  "description": "詳細な説明...",
  "severity": "high",
  "status": "open",
  "impact_scope": "Production API",
  "detected_at": "2024-01-01T10:00:00Z",
  "assignee_id": 2,
  "tag_ids": [1, 2]
}
```

**レスポンス** (201 Created):
```json
{
  "id": 1,
  "title": "Database connection timeout",
  "description": "詳細な説明...",
  "summary": "Database connection timeout occurred...",
  "severity": "high",
  "status": "open",
  "impact_scope": "Production API",
  "detected_at": "2024-01-01T10:00:00Z",
  "resolved_at": null,
  "assignee_id": 2,
  "creator_id": 1,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### 3.4 インシデント更新
**エンドポイント**: `PUT /api/incidents/:id`

**権限**: 編集者以上（自分の作成したインシデントのみ編集可能、管理者は全件編集可能）

**リクエスト**:
```json
{
  "title": "Database connection timeout (Updated)",
  "description": "更新された説明...",
  "severity": "critical",
  "status": "investigating",
  "impact_scope": "Production API and Web",
  "resolved_at": "2024-01-01T11:30:00Z",
  "assignee_id": 3,
  "tag_ids": [1, 3]
}
```

**レスポンス** (200 OK):
```json
{
  "id": 1,
  "title": "Database connection timeout (Updated)",
  "description": "更新された説明...",
  "summary": "Database connection timeout occurred...",
  "severity": "critical",
  "status": "investigating",
  "impact_scope": "Production API and Web",
  "detected_at": "2024-01-01T10:00:00Z",
  "resolved_at": "2024-01-01T11:30:00Z",
  "assignee_id": 3,
  "creator_id": 1,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T11:00:00Z"
}
```

### 3.5 インシデント削除
**エンドポイント**: `DELETE /api/incidents/:id`

**権限**: 管理者のみ

**レスポンス** (200 OK):
```json
{
  "message": "Incident deleted successfully"
}
```

### 3.6 AI要約生成
**エンドポイント**: `POST /api/incidents/:id/summarize`

**権限**: 編集者以上

**レスポンス** (200 OK):
```json
{
  "summary": "Database connection timeout occurred in production API. The issue was resolved within 1.5 hours.",
  "generated_at": "2024-01-01T10:05:00Z"
}
```

---

## 4. タイムラインAPI

### 4.1 タイムラインイベント追加
**エンドポイント**: `POST /api/incidents/:id/timeline`

**権限**: 編集者以上

**リクエスト**:
```json
{
  "event_type": "investigation_started",
  "event_time": "2024-01-01T10:15:00Z",
  "description": "Started investigation into database connection issues"
}
```

**レスポンス** (201 Created):
```json
{
  "id": 1,
  "incident_id": 1,
  "event_type": "investigation_started",
  "event_time": "2024-01-01T10:15:00Z",
  "description": "Started investigation into database connection issues",
  "recorder": {
    "id": 2,
    "name": "John Doe"
  },
  "created_at": "2024-01-01T10:15:00Z",
  "updated_at": "2024-01-01T10:15:00Z"
}
```

### 4.2 タイムラインイベント更新
**エンドポイント**: `PUT /api/timeline/:id`

**権限**: 編集者以上

**リクエスト**:
```json
{
  "event_type": "investigation_started",
  "event_time": "2024-01-01T10:20:00Z",
  "description": "Updated description"
}
```

**レスポンス** (200 OK):
```json
{
  "id": 1,
  "incident_id": 1,
  "event_type": "investigation_started",
  "event_time": "2024-01-01T10:20:00Z",
  "description": "Updated description",
  "recorder": {
    "id": 2,
    "name": "John Doe"
  },
  "created_at": "2024-01-01T10:15:00Z",
  "updated_at": "2024-01-01T10:20:00Z"
}
```

### 4.3 タイムラインイベント削除
**エンドポイント**: `DELETE /api/timeline/:id`

**権限**: 編集者以上

**レスポンス** (200 OK):
```json
{
  "message": "Timeline event deleted successfully"
}
```

---

## 5. タグAPI

### 5.1 タグ一覧取得
**エンドポイント**: `GET /api/tags`

**レスポンス** (200 OK):
```json
{
  "tags": [
    {
      "id": 1,
      "name": "Database",
      "color": "#FF5733",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 5.2 タグ作成
**エンドポイント**: `POST /api/tags`

**権限**: 管理者のみ

**リクエスト**:
```json
{
  "name": "Security",
  "color": "#33FF57"
}
```

**レスポンス** (201 Created):
```json
{
  "id": 2,
  "name": "Security",
  "color": "#33FF57",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 5.3 タグ更新
**エンドポイント**: `PUT /api/tags/:id`

**権限**: 管理者のみ

**リクエスト**:
```json
{
  "name": "Security Incident",
  "color": "#3357FF"
}
```

**レスポンス** (200 OK):
```json
{
  "id": 2,
  "name": "Security Incident",
  "color": "#3357FF",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:30:00Z"
}
```

### 5.4 タグ削除
**エンドポイント**: `DELETE /api/tags/:id`

**権限**: 管理者のみ

**レスポンス** (200 OK):
```json
{
  "message": "Tag deleted successfully"
}
```

---

## 6. ダッシュボードAPI

### 6.1 ダッシュボード統計取得
**エンドポイント**: `GET /api/dashboard/stats`

**クエリパラメータ**:
- `period` (string, default: month): 期間 (day/week/month)

**レスポンス** (200 OK):
```json
{
  "total_incidents": 150,
  "period_incidents": 25,
  "severity_distribution": {
    "critical": 5,
    "high": 10,
    "medium": 8,
    "low": 2
  },
  "status_distribution": {
    "open": 3,
    "investigating": 5,
    "resolved": 15,
    "closed": 2
  },
  "trend": [
    {
      "date": "2024-01-01",
      "count": 3
    },
    {
      "date": "2024-01-02",
      "count": 5
    }
  ],
  "recent_incidents": [
    {
      "id": 150,
      "title": "Latest incident",
      "severity": "high",
      "status": "open",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

## 7. ポストモーテムAPI (Phase 2)

### 7.1 ポストモーテム取得
**エンドポイント**: `GET /api/incidents/:id/postmortem`

**レスポンス** (200 OK):
```json
{
  "id": 1,
  "incident_id": 1,
  "root_cause": "根本原因の説明...",
  "good_points": "対応の良かった点...",
  "improvement_points": "改善が必要だった点...",
  "prevention_measures": "再発防止策...",
  "action_items": [
    {
      "id": 1,
      "description": "Implement connection pool monitoring",
      "assignee": {
        "id": 2,
        "name": "John Doe"
      },
      "due_date": "2024-02-01T00:00:00Z",
      "status": "pending"
    }
  ],
  "created_at": "2024-01-02T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}
```

### 7.2 ポストモーテム作成
**エンドポイント**: `POST /api/incidents/:id/postmortem`

**権限**: 編集者以上

**リクエスト**:
```json
{
  "root_cause": "根本原因の説明...",
  "good_points": "対応の良かった点...",
  "improvement_points": "改善が必要だった点...",
  "prevention_measures": "再発防止策...",
  "action_items": [
    {
      "description": "Implement connection pool monitoring",
      "assignee_id": 2,
      "due_date": "2024-02-01T00:00:00Z"
    }
  ]
}
```

**レスポンス** (201 Created):
```json
{
  "id": 1,
  "incident_id": 1,
  "root_cause": "根本原因の説明...",
  "good_points": "対応の良かった点...",
  "improvement_points": "改善が必要だった点...",
  "prevention_measures": "再発防止策...",
  "action_items": [
    {
      "id": 1,
      "description": "Implement connection pool monitoring",
      "assignee": {
        "id": 2,
        "name": "John Doe"
      },
      "due_date": "2024-02-01T00:00:00Z",
      "status": "pending"
    }
  ],
  "created_at": "2024-01-02T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}
```

### 7.3 ポストモーテム更新
**エンドポイント**: `PUT /api/postmortems/:id`

**権限**: 編集者以上

**リクエスト**: 作成時と同様

**レスポンス** (200 OK): 作成時と同様

---

## 8. 検索API (Phase 2)

### 8.1 高度な検索
**エンドポイント**: `GET /api/search/incidents`

**クエリパラメータ**:
- `q` (string): 全文検索クエリ
- `severity` (string): 深刻度フィルタ
- `status` (string): ステータスフィルタ
- `tag_ids` (string, comma-separated): タグIDのリスト
- `date_from` (string, ISO 8601): 開始日時
- `date_to` (string, ISO 8601): 終了日時
- `assignee_id` (integer): 担当者ID
- `creator_id` (integer): 作成者ID
- `page` (integer, default: 1): ページ番号
- `limit` (integer, default: 20): 1ページあたりの件数

**レスポンス** (200 OK): インシデント一覧取得と同様

---

## 9. 統計API (Phase 2)

### 9.1 MTTR統計取得
**エンドポイント**: `GET /api/stats/mttr`

**クエリパラメータ**:
- `start_date` (string, ISO 8601): 開始日
- `end_date` (string, ISO 8601): 終了日

**レスポンス** (200 OK):
```json
{
  "mttr_hours": 2.5,
  "mttr_minutes": 150,
  "total_resolved": 50,
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  }
}
```

### 9.2 タグ別統計取得
**エンドポイント**: `GET /api/stats/by-tags`

**レスポンス** (200 OK):
```json
{
  "tag_statistics": [
    {
      "tag": {
        "id": 1,
        "name": "Database",
        "color": "#FF5733"
      },
      "count": 25,
      "percentage": 16.67
    }
  ]
}
```

---

## 10. PDF生成API (Phase 3)

### 10.1 インシデントPDF生成
**エンドポイント**: `GET /api/incidents/:id/pdf`

**レスポンス** (200 OK):
- Content-Type: `application/pdf`
- バイナリデータ

### 10.2 サマリーレポートPDF生成
**エンドポイント**: `GET /api/reports/pdf`

**クエリパラメータ**:
- `start_date` (string, ISO 8601): 開始日
- `end_date` (string, ISO 8601): 終了日

**レスポンス** (200 OK):
- Content-Type: `application/pdf`
- バイナリデータ

---

## 11. ファイルAPI (Phase 3)

### 11.1 ファイルアップロード
**エンドポイント**: `POST /api/incidents/:id/files`

**権限**: 編集者以上

**リクエスト**: multipart/form-data
- `file` (file, max: 50MB): アップロードするファイル

**レスポンス** (201 Created):
```json
{
  "id": 1,
  "incident_id": 1,
  "file_name": "error.log",
  "file_size": 1048576,
  "mime_type": "text/plain",
  "file_url": "/api/files/1/download",
  "created_at": "2024-01-01T12:00:00Z"
}
```

### 11.2 ファイル一覧取得
**エンドポイント**: `GET /api/incidents/:id/files`

**レスポンス** (200 OK):
```json
{
  "files": [
    {
      "id": 1,
      "file_name": "error.log",
      "file_size": 1048576,
      "mime_type": "text/plain",
      "file_url": "/api/files/1/download",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### 11.3 ファイルダウンロード
**エンドポイント**: `GET /api/files/:id/download`

**レスポンス** (200 OK):
- Content-Type: ファイルのMIMEタイプ
- Content-Disposition: attachment; filename="..."
- バイナリデータ

### 11.4 ファイル削除
**エンドポイント**: `DELETE /api/files/:id`

**権限**: 編集者以上

**レスポンス** (200 OK):
```json
{
  "message": "File deleted successfully"
}
```

---

## 12. イベントタイプ定義

### タイムラインイベントタイプ
- `detected`: 検知
- `investigation_started`: 調査開始
- `root_cause_identified`: 原因特定
- `mitigation`: 緩和
- `resolved`: 解決
- `other`: その他

### 深刻度
- `critical`: Critical
- `high`: High
- `medium`: Medium
- `low`: Low

### ステータス
- `open`: Open
- `investigating`: Investigating
- `resolved`: Resolved
- `closed`: Closed

### ユーザー役割
- `admin`: 管理者
- `editor`: 編集者
- `viewer`: 閲覧者

### アクションアイテムステータス
- `pending`: 未対応
- `in_progress`: 対応中
- `completed`: 完了
- `cancelled`: キャンセル

