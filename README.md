# Incidex

<div align="center">

![Incidex Logo](./incidex_full_logo.jpg)

**Modern Incident Management System for SRE, DevOps, and Development Teams**

[English](./README_EN.md) | [日本語](./README.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Next.js Version](https://img.shields.io/badge/Next.js-14+-000000?logo=next.js)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript)](https://www.typescriptlang.org/)

</div>

---

## 📖 概要

**Incidex**（インシデックス）は、組織内で発生したインシデントを記録・管理し、AI要約とポストモーテムを通じて継続的な改善を促進するオープンソースのインシデント管理システムです。

インシデント情報をインデックス化し、組織の知見として蓄積することで、類似インシデントの再発を防ぎ、チームの学習と改善を支援します。

### ✨ 主な特徴

- 🤖 **AI要約機能**: インシデント詳細から自動で概要を生成（OpenAI API / Claude API対応）
- 📊 **タイムライン管理**: インシデントの経緯を時系列で記録・可視化
- 🏷️ **タグ管理**: カラー付きタグによる柔軟な分類とフィルタリング
- 📈 **統計ダッシュボード**: インシデント傾向の可視化とMTTRなどの指標追跡
- 📎 **ファイル添付**: ログやスクリーンショットなどの関連ファイルを管理
- 🔍 **高度な検索**: PostgreSQL全文検索による高速な検索機能
- 📄 **PDFレポート生成**: 期間指定でのサマリーレポート自動生成
- 🔐 **セルフホスト対応**: Docker Composeで簡単セットアップ、データは組織内に保持
- 🌐 **多言語対応**: 日本語・英語のUI対応（予定）

### 🎯 ターゲットユーザー

- 中小規模の開発チーム・SREチーム（5-50名）
- セキュリティオペレーションチーム（SOC）
- IT部門・情報システム部門
- コスト重視でセルフホスティングを希望する組織
- データを外部SaaSに出したくない組織（金融機関、官公庁等）

---

## 🚀 クイックスタート

### 前提条件

- Docker 20.10+ および Docker Compose 2.0+
- または、Go 1.21+ と Node.js 18+ がインストールされていること

### Docker Composeを使用した起動（推奨）

```bash
# リポジトリのクローン
git clone https://github.com/your-org/incidex.git
cd incidex

# 環境変数のセットアップ
cp .env.example .env
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local

# 本番環境の場合は、.envファイルの値を必ず変更してください
# 詳細は SECURITY.md を参照してください

# アプリケーションの起動
make up

# または
docker-compose up -d
```

起動後、以下のURLでアクセスできます：

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **MinIO Console**: http://localhost:9090（デフォルト: `minioadmin` / `minioadmin`）

### ローカル開発環境での起動

#### Backend

```bash
cd backend
cp .env.example .env
go mod download
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

詳細なセットアップ手順は [ドキュメント](./docs/) を参照してください。

---

## 📋 機能一覧

### Phase 1: 基本機能（実装済み）

- ✅ **認証・ユーザー管理**
  - ユーザー登録・ログイン（JWT認証）
  - ロールベースアクセス制御（管理者/編集者/閲覧者）
  - パスワードハッシュ化（bcrypt）

- ✅ **インシデント管理**
  - インシデントの作成・編集・削除・一覧表示
  - 深刻度（Critical/High/Medium/Low）とステータス管理
  - ページネーションと検索・フィルタリング機能
  - SLA管理と違反追跡

- ✅ **AI要約機能**
  - インシデント作成時の自動要約生成
  - 手動での要約再生成
  - OpenAI API / Claude API対応

- ✅ **タイムライン機能**
  - インシデントに紐づく時系列イベントの記録
  - イベントタイプ（検知、調査開始、原因特定、緩和、解決等）
  - コメント機能

- ✅ **タグ管理**
  - タグの作成・編集・削除
  - カラー設定による視覚的な分類
  - タグによるフィルタリング

- ✅ **ダッシュボード**
  - インシデント件数推移（日別・週別・月別）
  - 深刻度別・ステータス別の分布グラフ
  - 最近のインシデント一覧

- ✅ **ファイル添付**
  - インシデントへのファイル添付（画像、PDF、ログ等）
  - MinIOによるオブジェクトストレージ管理
  - ファイルのダウンロード・削除

### Phase 2: 高度な機能（開発中）

- 🔄 **ポストモーテム機能**
  - 根本原因分析（Five Whysテンプレート）
  - アクションアイテム管理
  - AI支援による根本原因分析の提案

- 🔄 **高度な検索・フィルタリング**
  - PostgreSQL全文検索（日本語・英語対応）
  - 複数条件での絞り込み
  - 検索結果のRedisキャッシュ

- 🔄 **統計・分析**
  - MTTR（平均復旧時間）の計算・表示
  - カテゴリ別のインシデント傾向分析
  - 再発率のトラッキング

### Phase 3: レポート機能（予定）

- 📄 **PDF自動生成**
  - 単体インシデントのPDFレポート出力
  - 期間指定でのサマリーレポート生成
  - カスタマイズ可能なレポートテンプレート

---

## 🛠 技術スタック

### Backend

- **Language**: Go 1.21+
- **Framework**: [Gin Web Framework](https://gin-gonic.com/)
- **ORM**: [GORM](https://gorm.io/)
- **Architecture**: Clean Architecture（`domain` / `usecase` / `interface` / `infrastructure`）
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Storage**: MinIO（S3互換オブジェクトストレージ）
- **AI**: OpenAI API / Claude API

### Frontend

- **Framework**: [Next.js 14+](https://nextjs.org/)（App Router）
- **Language**: TypeScript 5+
- **Styling**: [TailwindCSS](https://tailwindcss.com/)
- **State Management**: React Context API

### Infrastructure

- **Containerization**: Docker & Docker Compose
- **Tooling**: Make（開発・起動コマンドの共通化）

---

## 📁 プロジェクト構成

```
incidex/
├── backend/                 # Go Backend
│   ├── cmd/
│   │   ├── server/         # メインサーバー
│   │   └── seed/          # データベースシードツール
│   ├── internal/
│   │   ├── config/         # 設定管理
│   │   ├── domain/         # ドメインエンティティ・リポジトリIF
│   │   ├── usecase/       # ビジネスロジック
│   │   ├── interface/     # HTTPハンドラ・ルータ
│   │   └── infrastructure/ # DB・ストレージ・AI等の実装
│   └── Dockerfile
├── frontend/               # Next.js Frontend
│   ├── src/
│   │   ├── app/           # App Router ページ
│   │   ├── components/    # Reactコンポーネント
│   │   ├── context/       # グローバル状態管理
│   │   ├── lib/           # APIクライアント等
│   │   └── types/         # TypeScript型定義
│   └── Dockerfile
├── docs/                   # ドキュメント
│   ├── 要件定義書.md
│   ├── api-specification.md
│   ├── database-schema.md
│   └── プロジェクト計画書.md
├── docker-compose.yml      # Docker Compose設定
├── Makefile               # 開発用コマンド
├── README.md              # このファイル（日本語）
├── README_EN.md           # English README
├── SECURITY.md            # セキュリティガイドライン
├── CONTRIBUTING.md        # コントリビューションガイド
└── LICENSE                # ライセンスファイル
```

---

## 📚 ドキュメント

詳細なドキュメントは [`docs/`](./docs/) ディレクトリにあります：

- [要件定義書](./docs/要件定義書.md) - 機能要件と非機能要件の詳細
- [API仕様書](./docs/api-specification.md) - REST APIの詳細仕様
- [データベーススキーマ](./docs/database-schema.md) - データベース設計
- [ER図](./docs/er-diagram.md) - エンティティ関係図
- [プロジェクト計画書](./docs/プロジェクト計画書.md) - プロジェクト全体の計画

---

## 🔐 セキュリティ

セキュリティに関する重要な情報は [`SECURITY.md`](./SECURITY.md) に記載されています。

**本番環境で使用する前に必ず一読してください。**

主な注意事項：

- 強力な `JWT_SECRET` の設定（最低32文字）
- データベースSSLの有効化
- MinIO認証情報の変更
- HTTPS/TLSの設定

セキュリティ脆弱性を発見した場合は、公開Issueではなく、プロジェクトメンテナーに直接連絡してください。

---

## 🤝 コントリビューション

Incidexへのコントリビューションを歓迎します！

コントリビューション方法の詳細は [`CONTRIBUTING.md`](./CONTRIBUTING.md) を参照してください。

### コントリビューションの種類

- 🐛 **バグ報告**: Issueで問題を報告
- 💡 **機能提案**: 新機能や改善案の提案
- 🔧 **コード改善**: Pull Requestでのコード改善
- 📝 **ドキュメント改善**: ドキュメントの改善や翻訳
- 🧪 **テスト追加**: テストカバレッジの向上

### 開発フロー

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. Pull Requestを作成

---

## 📝 ライセンス

このプロジェクトは [MIT License](./LICENSE) の下で公開されています。

---

## 🗺 ロードマップ

### Phase 1: MVP（基本機能）✅
- 認証・ユーザー管理
- インシデントCRUD・検索・フィルタリング
- タグ管理
- AI要約機能
- タイムライン機能
- ダッシュボード

### Phase 2: 運用高度化 🔄
- ポストモーテム機能
- 高度な検索・フィルタリング
- 統計・分析機能の拡張

### Phase 3: レポート機能 📅
- PDFレポート生成
- カスタマイズ可能なレポートテンプレート

### 将来の計画
- マルチテナント対応（SaaS化）
- Webhook通知
- Slack連携
- 多言語UI対応の拡張
- Kubernetes Operator

詳細は [プロジェクト計画書](./docs/プロジェクト計画書.md) を参照してください。

---

## 💬 サポート

### Issue報告

バグ報告や機能要望は [GitHub Issues](https://github.com/your-org/incidex/issues) で受け付けています。

### ディスカッション

一般的な質問や議論は [GitHub Discussions](https://github.com/your-org/incidex/discussions) で行えます。

### セキュリティ問題

セキュリティに関する問題は、公開Issueではなく、プロジェクトメンテナーに直接連絡してください。

---

## 🙏 謝辞

Incidexは以下のオープンソースプロジェクトに依存しています：

- [Gin](https://gin-gonic.com/) - Go Web Framework
- [GORM](https://gorm.io/) - Go ORM
- [Next.js](https://nextjs.org/) - React Framework
- [TailwindCSS](https://tailwindcss.com/) - CSS Framework
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Cache
- [MinIO](https://min.io/) - Object Storage

その他、すべての依存パッケージの開発者に感謝します。

---

## 📞 連絡先

- **GitHub**: [https://github.com/your-org/incidex](https://github.com/your-org/incidex)
- **Issues**: [https://github.com/your-org/incidex/issues](https://github.com/your-org/incidex/issues)
- **Discussions**: [https://github.com/your-org/incidex/discussions](https://github.com/your-org/incidex/discussions)

---

<div align="center">

**Made with ❤️ by the Incidex Team**

[⭐ Star us on GitHub](https://github.com/your-org/incidex) | [📖 Documentation](./docs/) | [🤝 Contribute](./CONTRIBUTING.md)

</div>
