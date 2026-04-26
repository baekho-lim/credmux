#!/usr/bin/env bash
# credmux 데모 실행 스크립트
# 1분 데모 영상 순서대로: 평문 노출 → breach-drill → Telegram 안내.
#
# 사용 전: bash demo/setup.sh 로 더미 토큰 생성 + go build -o credmux .

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== Step 1: 평문 토큰 노출 ==="
cat demo/shop/.env | grep TOKEN

echo ""
echo "=== Step 2: breach-drill ==="
DEMO_HOME=./demo ./credmux breach-drill vercel

echo ""
echo "=== Step 3: Telegram bot은 스마트폰 확인 ==="
echo "→ /testbreach 입력"
