#!/bin/bash

# 統合テストスクリプト
# 新しいディレクトリ構造での動作確認

set -e

echo "🚀 Mongene 統合テスト開始"

# 色付きログ関数
log_info() {
    echo -e "\033[34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[31m[ERROR]\033[0m $1"
}

log_warning() {
    echo -e "\033[33m[WARNING]\033[0m $1"
}

# Docker Composeが利用可能かチェック
if ! command -v docker &> /dev/null; then
    log_error "docker が見つかりません。インストールしてください。"
    exit 1
fi

# 既存のコンテナを停止・削除
log_info "既存のコンテナを停止・削除中..."
docker compose down --remove-orphans || true

# コンテナをビルド・起動
log_info "コンテナをビルド・起動中..."
docker compose up --build -d

# サービスの起動を待機
log_info "サービスの起動を待機中..."
sleep 30

# ヘルスチェック
log_info "ヘルスチェック実行中..."

# backコンテナのヘルスチェック
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_success "backコンテナ: 正常"
else
    log_error "backコンテナ: 異常"
    docker compose logs back
    exit 1
fi

# coreコンテナのヘルスチェック
if curl -f http://localhost:1234/health > /dev/null 2>&1; then
    log_success "coreコンテナ: 正常"
else
    log_error "coreコンテナ: 異常"
    docker compose logs core
    exit 1
fi

# frontコンテナのヘルスチェック
if curl -f http://localhost:3000 > /dev/null 2>&1; then
    log_success "frontコンテナ: 正常"
else
    log_warning "frontコンテナ: 確認が必要（開発サーバーの起動に時間がかかる場合があります）"
fi

# API エンドポイントのテスト
log_info "API エンドポイントのテスト実行中..."

# 認証API テスト
log_info "認証API テスト..."
AUTH_RESPONSE=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"schoolCode":"00000","password":"password","remember":false}')

if echo "$AUTH_RESPONSE" | grep -q '"success":true'; then
    log_success "認証API: 正常"
else
    log_error "認証API: 異常"
    echo "Response: $AUTH_RESPONSE"
fi

# Core API テスト
log_info "Core API テスト..."
CORE_RESPONSE=$(curl -s http://localhost:1234/)

if echo "$CORE_RESPONSE" | grep -q "Mongene Core API Server"; then
    log_success "Core API: 正常"
else
    log_error "Core API: 異常"
    echo "Response: $CORE_RESPONSE"
fi

# ログの確認
log_info "コンテナログの確認..."
echo "=== Back Container Logs ==="
docker compose logs --tail=10 back

echo "=== Core Container Logs ==="
docker compose logs --tail=10 core

echo "=== Front Container Logs ==="
docker compose logs --tail=10 front

log_success "統合テスト完了！"
log_info "サービスは引き続き実行中です。停止するには 'docker compose down' を実行してください。"
log_info "アクセス先:"
log_info "  - Frontend: http://localhost:3000"
log_info "  - Backend API: http://localhost:8080"
log_info "  - Core API: http://localhost:1234"
log_info "  - Core API Docs: http://localhost:1234/docs"
