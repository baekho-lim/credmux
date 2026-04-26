#!/usr/bin/env bash
# demo/setup.sh — credmux 데모 환경 구성
#
# 5개 플랫폼(vercel/github/stripe/anthropic/supabase)의 더미 토큰을
# demo/{shop,blog,api}/ 아래 .env 파일로 흩뿌린다.
# 모든 토큰은 패턴만 진짜 같지 실제로는 무효 — 안전하게 시연 가능.
#
# 토큰은 prefix/suffix로 쪼개 변수에 보관 → 이 스크립트 자체가
# TruffleHog detector에 잡히지 않도록 한다.
#
# 매칭 현황 (TruffleHog 3.95.2 기준, 2026-04-26):
#   Vercel    ✅ 매칭됨
#   GitHub    ✅ 매칭됨
#   Stripe    ❌ detector regex와 더미 형식 불일치 (조사 후 패턴 보강 예정)
#   Anthropic ❌ detector regex와 더미 형식 불일치
#   Supabase  ❌ detector 자체가 trufflehog 3.95.2에 없음
# → 데모는 Vercel/GitHub로 진행. 위 3개는 혼란 방지를 위해 .env에 남겨두되
#   sentinel/breach-drill 시연 대상에서 제외 권장.
#
# Usage:
#   bash demo/setup.sh
#   DEMO_HOME=./demo ./credmux breach-drill vercel

set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "📦 building demo environment at $DEMO_DIR"

mkdir -p "$DEMO_DIR/shop" "$DEMO_DIR/blog" "$DEMO_DIR/api"

# Vercel API token: 24-char alnum + "vercel" keyword on same line.
V1_PRE='mUyAuxvD'; V1_SUF='3aIDybuGIKCKPxHu'
V2_PRE='qXjK9rN2'; V2_SUF='zPL4mEv8tBfYwShc'

# GitHub PAT: ghp_ + 36-char alnum.
GH_PRE='ghp_';     GH_SUF='K9mNpQ8vRjL4wYc9sT3bA5eF0hG6iD1uZX2YW'

# Stripe secret: sk_live_ + 99-char alnum (TruffleHog requires the full length).
SK_PRE='sk_live_'; SK_SUF='51AbCdEfGhIjKlMnOpQrStUvWxYz1234567890AbCdEfGhIjKlMnOpQrStUvWxYz1234567890ABCDEFGHIJKLMNOPQRSTUVWXY'

# Anthropic API key: sk-ant-apiNN- + 93 [\w-] chars + "AA" suffix (95 total tail).
AK_PRE='sk-ant-api03-'; AK_SUF='AbCdEfGhIjKlMnOpQrStUvWxYz1234567890_AbCdEfGhIjKlMnOpQrStUvWxYz1234567890_AbCdEfGhIjKlMnOpQrSAA'

# Supabase service-role key: JWT shape (eyJ... . eyJ... . sig). Best-effort.
SB_HDR='eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9'
SB_PAY='eyJzdWIiOiJkZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTcwMDAwMDAwMH0'
SB_SIG='dummy_signature_xxxxxxxxxxxxxxxxxxxxxxxxx'

cat > "$DEMO_DIR/shop/.env" <<EOF
# demo dummy — NOT real tokens
DATABASE_URL=postgresql://user:pass@localhost/shop
VERCEL_TOKEN=${V1_PRE}${V1_SUF}
STRIPE_SECRET=${SK_PRE}${SK_SUF}
NEXT_PUBLIC_API=https://api.example.com
EOF

cat > "$DEMO_DIR/blog/.env.local" <<EOF
# demo dummy — NOT real tokens
VERCEL_TOKEN=${V2_PRE}${V2_SUF}
GITHUB_TOKEN=${GH_PRE}${GH_SUF}
NODE_ENV=development
EOF

cat > "$DEMO_DIR/api/.env" <<EOF
# demo dummy — NOT real tokens
ANTHROPIC_API_KEY=${AK_PRE}${AK_SUF}
SUPABASE_URL=https://demo.supabase.co
SUPABASE_KEY=${SB_HDR}.${SB_PAY}.${SB_SIG}
EOF

echo "✅ created:"
find "$DEMO_DIR" -type f \( -name '.env' -o -name '.env.local' -o -name '.env.*' \) \
  | sort | sed "s|$DEMO_DIR/|  - |"

cat <<EOF

▶ verified to match in TruffleHog 3.95.2:
  DEMO_HOME=$DEMO_DIR ./credmux breach-drill vercel
  DEMO_HOME=$DEMO_DIR ./credmux breach-drill github

▶ stripe/anthropic/supabase: dummies present but currently unmatched
  (see CHECKLIST-CANDIDATES.md)
EOF
