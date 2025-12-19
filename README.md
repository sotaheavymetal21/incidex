# Incidex - Modern Incident Management System

Incidex は、高速でモダンなインシデント管理システムです。Go (Backend) と Next.js (Frontend) を使用し、堅牢な Clean Architecture に基づいて設計されています。

## ✨ 特徴

*   **認証システム**: JWTを使用したセキュアなログイン・サインアップ機能。
*   **タグ管理**: インシデントを整理するための柔軟なタグ付けシステム。
*   **インシデント管理**: インシデントの作成、追跡、ステータス管理 (開発中)。
*   **タイムライン**: インシデントの経緯を時系列で可視化 (開発中)。
*   **モダンなUI**: Next.js と TailwindCSS によるレスポンシブで使いやすいインターフェース。

## 🛠 技術スタック

### Backend
*   **Language**: Go 1.21+
*   **Framework**: Gin Web Framework
*   **ORM**: GORM
*   **Architecture**: Clean Architecture (Domain, Usecase, Infrastructure, Interface)
*   **Database**: PostgreSQL
*   **Cache**: Redis
*   **Storage**: MinIO (S3 Compatible)

### Frontend
*   **Framework**: Next.js 14+ (App Router)
*   **Language**: TypeScript
*   **Styling**: TailwindCSS
*   **State Management**: React Context API

### Infrastructure
*   **Containerization**: Docker & Docker Compose
*   **Tooling**: Make

## 🚀 環境構築 (Getting Started)

### 前提条件
*   Docker & Docker Compose がインストールされていること。

### セットアップ手順

1.  **リポジトリのクローン**
    ```bash
    git clone <repository-url>
    cd incidex
    ```

2.  **環境変数の設定**

    **重要**: 本番環境にデプロイする前に、必ず [SECURITY.md](./SECURITY.md) を確認してください。

    ```bash
    # ルートディレクトリ（Docker Compose用）
    cp .env.example .env

    # バックエンド
    cp backend/.env.example backend/.env

    # フロントエンド
    cp frontend/.env.example frontend/.env.local
    ```

    開発環境では、デフォルト値をそのまま使用できます。

    **本番環境では必ず以下を変更してください**:
    - `JWT_SECRET`: 強力なランダム文字列（32文字以上）
    - `POSTGRES_PASSWORD`: 強力なパスワード
    - `MINIO_ROOT_PASSWORD`: 強力なパスワード
    - `APP_ENV`: `production` に設定

    詳細は [SECURITY.md](./SECURITY.md) を参照してください。

3.  **アプリケーションの起動**
    付属の `Makefile` を使用して簡単に起動できます。

    ```bash
    make up
    ```
    このコマンドで、PostgreSQL, Redis, MinIO, Backend, Frontend の全てのコンテナが起動します。

4.  **アクセス**
    *   **Frontend**: [http://localhost:3000](http://localhost:3000)
    *   **Backend API**: [http://localhost:8080](http://localhost:8080)
    *   **MinIO Console**: [http://localhost:9090](http://localhost:9090) (User: `minioadmin`, Pass: `minioadmin`)

### 開発コマンド

*   `make up`: 全サービスの起動 (バックグラウンド)
*   `make down`: 全サービスの停止・削除
*   `make logs`: ログの表示
*   `make restart`: 再起動
*   `make dev`: ローカル開発モード (Go と Next.js をローカルで動かす場合)

## 📂 ディレクトリ構造

```
incidex/
├── backend/            # Go Backend
│   ├── cmd/            # Entry points
│   ├── internal/       # Application code
│   │   ├── config/     # Configuration
│   │   ├── domain/     # Enterprise Business Rules (Entities)
│   │   ├── usecase/    # Application Business Rules
│   │   ├── interface/  # Interface Adapters (Handlers, Routers)
│   │   └── infrastructure/ # Frameworks & Drivers (DB, External APIs)
│   └── go.mod
├── frontend/           # Next.js Frontend
│   ├── src/
│   │   ├── app/        # App Router Pages
│   │   ├── components/ # Reusable Components
│   │   ├── context/    # Global State (Auth etc.)
│   │   ├── lib/        # Utilities (API Client etc.)
│   │   └── types/      # TypeScript Definitions
│   └── package.json
├── docker-compose.yml  # Docker services definition
├── Makefile            # Development commands
└── docs/               # Documentation (Requirements, Schema, etc.)
```

## 📖 ドキュメント

詳細なドキュメントは `docs/` ディレクトリを参照してください。
*   [要件定義書](docs/要件定義書.md)
*   [API仕様書](docs/api-specification.md)
*   [データベース設計](docs/database-schema.md)

## 📝 License

Checking...