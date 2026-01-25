#!/bin/sh
set -e

# AUTO_MIGRATE が有効な場合、データベースマイグレーションを実行します
if [ "$AUTO_MIGRATE" = "true" ]; then
    echo "INFO: AUTO_MIGRATE が有効です。データベースマイグレーションを実行します..."

    # データベースの準備ができるまで待機します
    until goose -dir "${MIGRATIONS_DIR:-/root/migrations}" postgres "$DATABASE_URL" version > /dev/null 2>&1; do
        echo "データベースの準備を待っています..."
        sleep 2
    done

    # マイグレーションを実行します
    echo "goose マイグレーションを実行しています..."
    goose -dir "${MIGRATIONS_DIR:-/root/migrations}" postgres "$DATABASE_URL" up

    echo "SUCCESS: データベースマイグレーションが完了しました"
else
    echo "INFO: AUTO_MIGRATE が無効です。マイグレーションをスキップします。"
fi

# メインアプリケーションを起動します
echo "アプリケーションを起動しています..."
exec "$@"
