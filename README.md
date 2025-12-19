## Incidex - Modern Incident Management System

Incidex は、**SRE / DevOps / 開発チーム向けのモダンなインシデント管理システム**です。  
Go 製バックエンドと Next.js 製フロントエンドで構成され、**クリーンアーキテクチャ**と **コンテナベースのインフラ**を採用しています。

将来的には **SaaS 提供**と **OSS としての一般公開**の両立を目指しており、  
「セルフホストでも、マネージドでも使えるインシデント管理プラットフォーム」をゴールとしています。

---

### ✨ コンセプトと特徴

- **インシデントライフサイクルの一元管理**  
  - 登録 → 調査 → 対応 → 解決 → ポストモーテムまでを一つのタイムラインで管理
- **SaaS / OSS を見据えたアーキテクチャ**  
  - クリーンアーキテクチャにより、クラウド / オンプレ / マルチテナント構成への拡張を意識
- **チームのナレッジ蓄積を支援**  
  - AI 要約、タイムライン、タグ、ポストモーテム機能（Phase 2 以降）で「再発しないための知識」を残す
- **セルフホストしやすい設計**  
  - Docker Compose + Makefile で、ワンコマンド起動を前提にした構成

#### 現在実装済み / 開発中の主な機能

- **認証・認可**
  - JWT ベースのサインアップ・ログイン
  - ロール（`admin` / `editor` / `viewer`）による権限制御
- **タグ管理**
  - インシデントに紐づくタグの作成・編集・削除
  - 色付きタグによる分類とフィルタリング
- **インシデント管理（Phase 1 実装中）**
  - インシデントの作成・編集・削除
  - 深刻度・ステータス・影響範囲・担当者などの属性管理
  - 一覧・詳細閲覧、フィルタリング・検索
- **タイムライン / ポストモーテム / 統計（Phase 2 以降）**
  - タイムラインイベント、ポストモーテム、ダッシュボード、PDF 出力などを順次実装予定

詳細な要件は `docs/要件定義書.md` を参照してください。

---

### 🧭 SaaS / OSS としての方向性

- **現時点**
  - 主にセルフホストを想定した開発中プロジェクトです。
  - 個人 / 小規模チームのインシデント管理・学習用途での利用を想定しています。
- **中期的なロードマップ**
  - マルチテナント対応や監査ログなど、SaaS 提供に必要な機能の追加
  - OSS として公開し、コントリビューションを受け入れられる体制の整備
- **長期ビジョン**
  - Incidex をベースにしたマネージド SaaS 提供
  - コアは OSS として公開しつつ、エンタープライズ拡張や運用支援ツールをオプション提供

OSS 化後は、Issue / Discussion / Pull Request を通じて幅広いコントリビューションを歓迎する予定です。

---

### 🛠 技術スタック

#### Backend
- **Language**: Go 1.21+
- **Framework**: Gin Web Framework
- **ORM**: GORM
- **Architecture**: Clean Architecture（`domain` / `usecase` / `interface` / `infrastructure`）
- **Database**: PostgreSQL
- **Cache**: Redis（統計・AI 要約などのキャッシュ用途を想定）
- **Storage**: MinIO（S3 互換オブジェクトストレージ、ファイル添付など Phase 3 で利用予定）

#### Frontend
- **Framework**: Next.js 14+（App Router）
- **Language**: TypeScript
- **Styling**: TailwindCSS
- **State Management**: React Context API（`AuthContext` など）

#### Infrastructure
- **Containerization**: Docker & Docker Compose
- **Tooling**: Make（開発・起動コマンドを共通化）

---

### 🚀 クイックスタート（Docker 利用）

#### 前提条件

- Docker / Docker Compose がインストールされていること

#### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd incidex
```

#### 2. 環境変数のセットアップ

**本番環境での利用・インターネットに公開する場合は、必ず先に [SECURITY.md](./SECURITY.md) を確認してください。**

```bash
# ルートディレクトリ（Docker Compose 用）
cp .env.example .env

# バックエンド
cp backend/.env.example backend/.env

# フロントエンド
cp frontend/.env.example frontend/.env.local
```

- 開発用途であれば、`.env.example` のデフォルト値で動作します。
- 本番 / インターネット公開時は、**以下の値を必ず変更してください**。
  - `JWT_SECRET`（32 文字以上の十分に強いランダム文字列）
  - Database / Redis / MinIO などの各種パスワード・シークレット
  - `APP_ENV=production`

より詳細な項目や推奨値は `SECURITY.md` を参照してください。

#### 3. アプリケーションの起動

```bash
make up
```

このコマンドで PostgreSQL / Redis / MinIO / Backend / Frontend が一括で起動します。

#### 4. アクセス

- **Frontend**: `http://localhost:3000`
- **Backend API**: `http://localhost:8080`
- **MinIO Console**: `http://localhost:9090`（初期値: User `minioadmin`, Password `minioadmin`）

---

### 🔧 ローカル開発（Docker を使わない場合）

Backend / Frontend を個別に立ち上げたい場合の手順です。

#### Backend（Go）

```bash
cd backend
cp .env.example .env  # 未作成の場合
go run cmd/server/main.go
```

- デフォルトではポート `8080` で起動します。
- DB などの接続先は `.env` の `DATABASE_URL` などで調整してください。

#### Frontend（Next.js）

```bash
cd frontend
cp .env.example .env.local  # 未作成の場合
npm install
npm run dev
```

- `http://localhost:3000` で起動します。
- `NEXT_PUBLIC_API_URL`（例: `http://localhost:8080/api`）がバックエンドに向くように設定してください。

#### Makefile でのサポートコマンド

- `make up`：全サービスの Docker コンテナ起動
- `make down`：コンテナの停止・削除
- `make logs`：すべてのコンテナのログ表示
- `make restart`：再起動
- `make dev`：バックエンド / フロントエンドをローカル実行（開発向け）

---

### 🏗 アーキテクチャ概要

#### Backend（Clean Architecture）

- `internal/domain`  
  - `User` / `Tag` / `Incident` などのドメインエンティティ
  - `UserRepository` などのリポジトリインターフェース
- `internal/usecase`  
  - `AuthUsecase` / `TagUsecase` / `IncidentUsecase` などアプリケーションロジック
- `internal/interface/http`  
  - Gin のハンドラー / ルーター / 認証ミドルウェア
- `internal/infrastructure/persistence`  
  - GORM を利用した DB リポジトリ実装

この構成により、**UI やインフラに依存しないビジネスロジック**を維持しつつ、SaaS 化や他プロトコル（gRPC など）への拡張がしやすくなっています。

#### Frontend（App Router 構成）

- `src/app`  
  - `/login` / `/signup` / `/tags` / `/incidents` などのページ
- `src/context`  
  - `AuthContext` によるログイン状態・トークン管理
- `src/lib/api.ts`  
  - 共通 API クライアント、`authApi` / `tagApi` など
- `src/types`  
  - バックエンド API と対応する TypeScript 型

---

### 📂 リポジトリ構成（概要）

```
incidex/
├── backend/            # Go Backend
│   ├── cmd/            # エントリポイント
│   ├── internal/       # アプリケーションコード
│   │   ├── config/     # 設定読み込み
│   │   ├── domain/     # エンティティ・リポジトリIF
│   │   ├── usecase/    # ビジネスロジック
│   │   ├── interface/  # HTTPハンドラ・ルータ
│   │   └── infrastructure/ # DB など外部依存
├── frontend/           # Next.js Frontend
│   ├── src/
│   │   ├── app/        # App Router ページ
│   │   ├── context/    # グローバル状態（認証など）
│   │   ├── lib/        # API クライアント等
│   │   └── types/      # 型定義
├── docker-compose.yml  # 各種サービス定義
├── Makefile            # 開発用コマンド
└── docs/               # 仕様・設計ドキュメント
```

---

### 📖 ドキュメント

より詳細な仕様・設計は `docs/` 以下にまとまっています。

- **要件定義**: `docs/要件定義書.md`
- **API 仕様**: `docs/api-specification.md`
- **DB スキーマ / ER 図**: `docs/database-schema.md`, `docs/er-diagram.md`
- **プロジェクト計画**: `docs/プロジェクト計画書.md`

---

### 🤝 コントリビューション（予定方針）

Incidex は将来的に OSS として一般公開し、外部からのコントリビューションを積極的に受け入れることを目指しています。  
現時点ではクローズドな開発フェーズですが、以下のような貢献を歓迎する予定です。

- 機能提案・ Issue 作成（新機能 / 不具合報告 / 改善案）
- バグ修正・リファクタリング
- ドキュメントの改善・翻訳
- SaaS 運用を見据えた監視・運用プラクティスの共有

公開後の基本的な方針（予定）:

- **Pull Request ベースの開発フロー**
- Go / TypeScript の標準的な Lint / Format に準拠
- 可能な範囲でユニットテスト / E2E テストの追加を推奨

具体的なガイドラインは OSS 公開時に `CONTRIBUTING.md` として整備する予定です。

---

### 🔐 セキュリティと責任ある開示

セキュリティに関する設計・運用の詳細は `SECURITY.md` にまとめています。

- 本番環境での利用前に **必ず** 一読してください。
- セキュリティインシデントを発見した場合は、公開 Issue ではなく、**メンテナーへの直接連絡**を推奨しています。

---

### 🗺 ロードマップ（抜粋）

**Phase 1（基本機能）**
- 認証 / ユーザー管理（ロールベースアクセス制御）
- インシデント CRUD・検索・フィルタリング
- タグ管理

**Phase 2（運用高度化）**
- タイムライン
- ポストモーテム
- 高度な検索・統計（MTTR, タグ別集計など）

**Phase 3（レポーティング・ファイル連携）**
- PDF レポート生成
- ファイル添付（MinIO 連携）
- より高度なダッシュボード・レポーティング

詳細は `docs/要件定義書.md` および `docs/api-specification.md` を参照してください。

---

### 📝 ライセンス

現在、Incidex のライセンス形態は **検討中** です。  
OSS として一般公開するタイミングで、適切な OSS ライセンス（例: MIT / Apache-2.0 など）を明示します。

それまでは、事前の合意なく商用サービスとして再配布・再販売することはお控えください。
