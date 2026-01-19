#!/bin/bash
#
# 監査ログのクリーンアップスクリプト
# 30日以上前のログを削除する
#
# 使い方:
#   ./scripts/cleanup-audit-logs.sh                    # docker経由で実行
#   ./scripts/cleanup-audit-logs.sh --direct           # 直接psql実行（本番用）
#   RETENTION_DAYS=60 ./scripts/cleanup-audit-logs.sh  # 保持日数を変更
#

set -e

RETENTION_DAYS=${RETENTION_DAYS:-30}

echo "Cleaning up audit logs older than ${RETENTION_DAYS} days..."

if [ "$1" = "--direct" ]; then
    # 本番環境向け: 環境変数からDB接続情報を取得
    if [ -z "$DATABASE_URL" ]; then
        echo "Error: DATABASE_URL environment variable is required for --direct mode"
        exit 1
    fi
    psql "$DATABASE_URL" -c "DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '${RETENTION_DAYS} days'"
else
    # 開発環境向け: docker-compose経由で実行
    docker compose exec -T db psql -U user -d incidex -c "DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '${RETENTION_DAYS} days'"
fi

echo "Cleanup completed."
