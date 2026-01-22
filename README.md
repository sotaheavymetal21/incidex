# Incidex

![Incidex Logo](./incidex_full_logo.jpg)

**組織内で発生したインシデントを記録・管理し、継続的な改善を促進するオープンソースのインシデント管理システム**

[English](./README_EN.md) | [日本語](./README.md)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![Next.js Version](https://img.shields.io/badge/Next.js-16+-000000?logo=next.js)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript)](https://www.typescriptlang.org/)

---

## 概要

Incidexは、インシデント情報を記録・分類し、組織の知見として蓄積するためのシステムです。タイムライン管理、AI要約、ポストモーテム、統計分析などの機能により、チームの学習と改善を支援します。

### 主な特徴

- インシデントの作成・編集・検索とステータス管理
- タイムライン機能による時系列イベントの記録
- ファイル添付（ログ、スクリーンショット等）
- タグによる分類とフィルタリング
- ポストモーテムとアクションアイテム管理
- 統計ダッシュボードとSLA追跡
- レポート生成（月次レポート、カスタムレポート）
- PDFエクスポート（インシデント詳細、CSV一括出力）
- 監査ログによる操作履歴の記録
- 通知設定（インシデント作成、コメント追加等）
- ロールベースアクセス制御（管理者/編集者/閲覧者）
- セルフホスト対応（Docker Compose）

### 想定ユーザー

- 開発チーム、SREチーム
- セキュリティオペレーションチーム
- IT部門、情報システム部門
- セルフホスティングを希望する組織

---

## クイックスタート

### 前提条件

- Docker 20.10+ および Docker Compose 2.0+
- または、Go 1.24+ と Node.js 18+ がインストール済みであること

### Docker Composeでの起動

```bash
git clone https://github.com/your-org/incidex.git
cd incidex

# 環境変数のセットアップ
cp backend/.env.example backend/.env
# 本番環境では .env ファイルの値を変更してください（SECURITY.md 参照）

# アプリケーションの起動
make up
```

起動後のアクセス先:

- Frontend: <http://localhost:3000>
- Backend API: <http://localhost:8080>
- MinIO Console: <http://localhost:9090>

初回起動時は、環境変数で指定した管理者アカウントが自動作成されます（デフォルト: <admin@example.com> / admin123）。

### ローカル開発環境での起動

```bash
# Backend
cd backend
cp .env.example .env
go mod download
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm run dev
```

---

## 機能一覧

### 認証・ユーザー管理

- ユーザー登録・ログイン（JWT認証）
- パスワードリセット機能（メール送信対応）
- ロールベースアクセス制御（管理者/編集者/閲覧者）
- 管理者によるユーザー管理（作成、更新、無効化、削除）

### インシデント管理

- インシデントの作成・編集・削除・一覧表示
- 深刻度（Critical/High/Medium/Low）とステータス管理
- インシデントの担当者割り当て
- ページネーションと検索・フィルタリング
- SLA管理と違反追跡

### タイムライン・アクティビティ

- タイムラインイベントの記録（検知、調査開始、原因特定、緩和、解決等）
- コメント機能
- アクティビティ履歴の表示

### ファイル管理

- インシデントへのファイル添付（画像、PDF、ログ等）
- MinIOによるオブジェクトストレージ管理
- ファイルのダウンロード・削除

### タグ管理

- タグの作成・編集・削除
- カラー設定による分類
- タグによるフィルタリング

### ポストモーテム

- ポストモーテムの作成・編集・削除
- 公開/非公開ステータス管理
- アクションアイテムの登録・管理
- 根本原因分析の記録

### 統計・ダッシュボード

- インシデント件数推移（日別・週別・月別）
- 深刻度別・ステータス別の分布
- SLAメトリクス
- タグ別統計
- 最近のインシデント一覧

### レポート・エクスポート

- 月次レポート生成
- カスタム期間でのレポート生成
- インシデント詳細のPDFエクスポート
- インシデント一覧のCSVエクスポート

### 監査ログ

- 全ての操作履歴の記録
- ユーザー、リソース、操作種別による検索

### 通知設定

- インシデント作成時の通知
- コメント追加時の通知
- タイムラインイベント追加時の通知
- ユーザーごとの通知設定

---

## 技術スタック

### Backend

- Language: Go 1.24
- Framework: Gin Web Framework
- ORM: GORM
- Architecture: Clean Architecture (domain / usecase / interface / infrastructure)
- Database: PostgreSQL 15
- Cache: Redis 7
- Storage: MinIO (S3互換オブジェクトストレージ)
- Migration: goose

### Frontend

- Framework: Next.js 16 (App Router)
- Language: TypeScript 5
- Runtime: React 19
- Styling: TailwindCSS 4
- Charts: Recharts

### Infrastructure

- Containerization: Docker & Docker Compose
- Development Tools: Make

---

## プロジェクト構成

```
incidex/
├── backend/
│   ├── cmd/
│   │   ├── server/         # メインサーバー
│   │   └── seed/           # データベースシードツール
│   ├── internal/
│   │   ├── config/         # 設定管理
│   │   ├── domain/         # ドメインエンティティ・リポジトリIF
│   │   ├── usecase/        # ビジネスロジック
│   │   ├── interface/      # HTTPハンドラ・ルータ
│   │   └── infrastructure/ # DB・ストレージ・通知等の実装
│   └── migrations/         # データベースマイグレーション
├── frontend/
│   └── src/
│       ├── app/            # App Routerページ
│       ├── components/     # Reactコンポーネント
│       ├── context/        # グローバル状態管理
│       ├── lib/            # APIクライアント
│       └── types/          # TypeScript型定義
├── docs/                   # ドキュメント
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## ドキュメント

詳細なドキュメントは `docs/` ディレクトリを参照してください。

- API仕様書
- データベーススキーマ
- 要件定義書

---

## データベースマイグレーション

データベースマイグレーションには goose を使用します。

### マイグレーション実行

Docker環境:

```bash
make migrate-docker-up       # マイグレーション実行
make migrate-docker-status   # ステータス確認
make migrate-docker-down     # ロールバック
```

ローカル開発環境:

```bash
make migrate-up       # マイグレーション実行
make migrate-status   # ステータス確認
make migrate-down     # ロールバック
```

### 新規マイグレーションの作成

```bash
make migrate-create name=add_new_feature
```

作成されたファイル（`backend/migrations/YYYYMMDDHHMMSS_add_new_feature.sql`）の `-- +goose Up` セクションにマイグレーション処理、`-- +goose Down` セクションにロールバック処理を記述します。

---

## セキュリティ

本番環境での使用前に `SECURITY.md` を確認してください。

主な注意事項:

- 強力なJWT_SECRETの設定（最低32文字）
- データベースSSLの有効化
- MinIO認証情報の変更
- HTTPS/TLSの設定

セキュリティ脆弱性の報告は、公開Issueではなくプロジェクトメンテナーに直接連絡してください。

---

## コントリビューション

プルリクエストやバグ報告を歓迎します。詳細は `CONTRIBUTING.md` を参照してください。

1. リポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/new-feature`)
3. 変更をコミット
4. ブランチにプッシュ
5. Pull Requestを作成
