# Incidex Security Guide

このドキュメントは、Incidexを安全にデプロイ・運用するためのセキュリティガイドラインです。

## 目次

1. [環境変数の設定](#環境変数の設定)
2. [本番環境デプロイメント](#本番環境デプロイメント)
3. [セキュリティチェックリスト](#セキュリティチェックリスト)
4. [既知の制限事項](#既知の制限事項)

---

## 環境変数の設定

### 開発環境のセットアップ

1. **ルートディレクトリ** (Docker Compose用)
   ```bash
   cp .env.example .env
   # .env ファイルを編集して適切な値を設定
   ```

2. **バックエンド**
   ```bash
   cd backend
   cp .env.example .env
   # .env ファイルを編集
   ```

3. **フロントエンド**
   ```bash
   cd frontend
   cp .env.example .env.local
   # .env.local ファイルを編集
   ```

### 必須環境変数

#### Backend (.env)

| 変数名 | 説明 | 開発環境のデフォルト | 本番環境の要件 |
|--------|------|---------------------|---------------|
| `JWT_SECRET` | JWT署名用の秘密鍵 | `super-secret-key-change-me-in-production` | **最低32文字の強力なランダム文字列** |
| `DATABASE_URL` | PostgreSQL接続文字列 | `postgres://user:password@localhost:5432/incidex?sslmode=disable` | **SSL有効化を推奨** (`sslmode=require`) |
| `REDIS_URL` | Redis接続文字列 | `localhost:6379` | 適切なホスト名を設定 |
| `MINIO_ACCESS_KEY` | MinIOアクセスキー | `minioadmin` | **強力なランダム文字列に変更** |
| `MINIO_SECRET_KEY` | MinIOシークレットキー | `minioadmin` | **強力なランダム文字列に変更** |
| `APP_ENV` | 実行環境 | `development` | **`production`に設定** |

#### Frontend (.env.local)

| 変数名 | 説明 | 開発環境のデフォルト | 本番環境の値 |
|--------|------|---------------------|-------------|
| `NEXT_PUBLIC_API_URL` | バックエンドAPI URL | `http://localhost:8080/api` | `https://your-domain.com/api` |

---

## 本番環境デプロイメント

### 1. 強力な秘密鍵の生成

#### JWT_SECRET

```bash
# OpenSSLを使用して32バイトのランダム文字列を生成（Base64エンコード）
openssl rand -base64 32
```

#### MinIO Credentials

```bash
# 20文字のランダム文字列を生成
openssl rand -hex 20
```

### 2. 本番環境チェック

バックエンドは `APP_ENV=production` が設定されている場合、起動時に以下をチェックします：

- ✅ JWT_SECRETが最低32文字であること
- ✅ デフォルト値が使用されていないこと
- ✅ MinIO認証情報がデフォルトから変更されていること
- ⚠️ Database SSLが無効化されている場合は警告

**チェックに失敗すると、アプリケーションは起動を拒否します。**

### 3. Database SSL設定

本番環境では、必ずSSLを有効化してください：

```bash
DATABASE_URL=postgres://user:password@host:5432/incidex?sslmode=require
```

### 4. HTTPS/TLS設定

- フロントエンドとバックエンドは必ずHTTPS経由で提供してください
- リバースプロキシ（Nginx, Caddy等）でTLS終端を行うことを推奨

---

## セキュリティチェックリスト

### デプロイ前

- [ ] すべての `.env` ファイルがバージョン管理から除外されている（`.gitignore`で確認）
- [ ] `JWT_SECRET` が強力なランダム文字列に設定されている
- [ ] `DATABASE_URL` にSSL設定が含まれている（`sslmode=require`）
- [ ] MinIO認証情報がデフォルトから変更されている
- [ ] PostgreSQLユーザーのパスワードが強力である
- [ ] `APP_ENV=production` が設定されている
- [ ] フロントエンドのAPI URLがHTTPSになっている

### 運用中

- [ ] 定期的にログを監視している
- [ ] データベースの定期バックアップが実施されている
- [ ] 依存パッケージの脆弱性スキャンを実施している
- [ ] JWTトークンの有効期限が適切に設定されている（現在24時間）
- [ ] 不要なポートが外部に公開されていない

---

## 既知の制限事項

### 認証とアクセス制御

1. **JWT有効期限**: 現在24時間に設定されています。要件に応じて調整してください。
2. **リフレッシュトークン**: 未実装。長時間セッションが必要な場合は実装を検討してください。
3. **Rate Limiting**: 未実装。DDoS対策として、リバースプロキシレベルでの実装を推奨します。
4. **CSRF保護**: 現在未実装。フロントエンドとバックエンドが異なるドメインの場合は実装を検討してください。

### データ保護

1. **暗号化**: データベース内のデータは暗号化されていません。機密性の高いデータを扱う場合は、データベース暗号化を検討してください。
2. **監査ログ**: 現在未実装。コンプライアンス要件がある場合は実装を検討してください。

### ネットワークセキュリティ

1. **CORS設定**: 適切なCORS設定を行ってください（現在は設定されていません）
2. **ファイアウォール**: 必要なポートのみを公開してください
   - Frontend: 3000 (開発), 443 (本番)
   - Backend: 8080 (開発), 443 (本番、リバースプロキシ経由)
   - PostgreSQL: 5432 (内部ネットワークのみ)
   - Redis: 6379 (内部ネットワークのみ)
   - MinIO: 9000, 9090 (内部ネットワークのみ)

---

## パスワードポリシー

### ユーザーパスワード

- 最小長: 6文字（現在の実装）
- **推奨**: 本番環境では最小8文字以上、複雑性要件の追加を検討

### システムパスワード（データベース、MinIO等）

- 最小長: 16文字以上
- 英数字 + 記号を含む
- 定期的な変更（90日ごと推奨）

---

## インシデント対応

セキュリティインシデントが発生した場合：

1. **即座に実施**
   - 影響を受けたサービスの隔離
   - 秘密鍵の無効化と再生成
   - すべてのユーザーセッションの無効化

2. **調査**
   - アクセスログの確認
   - データベースの整合性チェック
   - 影響範囲の特定

3. **復旧**
   - 脆弱性の修正
   - 新しい秘密鍵での再デプロイ
   - ユーザーへの通知（必要に応じて）

---

## セキュリティアップデート

### 依存パッケージの更新

#### Backend (Go)

```bash
cd backend
go get -u ./...
go mod tidy
```

#### Frontend (npm)

```bash
cd frontend
npm audit
npm audit fix
# または
npm update
```

### 定期的なチェック

- 週次: 脆弱性スキャン
- 月次: 依存パッケージの更新確認
- 四半期: セキュリティ設定の見直し

---

## 参考リソース

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://golang.org/doc/security/)
- [Next.js Security Headers](https://nextjs.org/docs/advanced-features/security-headers)
- [PostgreSQL Security](https://www.postgresql.org/docs/current/security.html)

---

## 連絡先

セキュリティに関する問題を発見した場合は、公開イシューではなく、プロジェクトメンテナーに直接連絡してください。

---

**最終更新**: 2024-12-19
