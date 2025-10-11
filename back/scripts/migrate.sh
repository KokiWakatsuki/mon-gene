#!/bin/bash

# データベースマイグレーションスクリプト

set -e

# 環境変数の読み込み
if [ -f ../.env ]; then
    export $(cat ../.env | grep -v '^#' | xargs)
fi

# デフォルト値
DB_HOST=${DB_HOST:-db}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-user}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${DB_NAME:-develop}

echo "📦 データベースマイグレーションを開始します..."
echo "接続先: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

# データベース接続の確認
echo "🔍 データベース接続を確認しています..."
until mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" -p"${DB_PASSWORD}" -e "SELECT 1" > /dev/null 2>&1; do
    echo "⏳ データベースの起動を待機中..."
    sleep 2
done

echo "✅ データベースに接続しました"

# マイグレーションファイルを実行
MIGRATION_DIR="../migrations"

if [ ! -d "$MIGRATION_DIR" ]; then
    echo "❌ エラー: マイグレーションディレクトリが見つかりません: $MIGRATION_DIR"
    exit 1
fi

# マイグレーションファイルを順番に実行
for file in $(ls $MIGRATION_DIR/*.sql | sort); do
    echo "📄 実行中: $(basename $file)"
    mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" -p"${DB_PASSWORD}" "${DB_NAME}" < "$file"
    if [ $? -eq 0 ]; then
        echo "✅ 完了: $(basename $file)"
    else
        echo "❌ エラー: $(basename $file)"
        exit 1
    fi
done

echo "🎉 マイグレーションが完了しました！"
