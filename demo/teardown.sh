#!/usr/bin/env bash
# demo/teardown.sh — 데모 환경 정리
# setup.sh로 만든 더미 .env 디렉토리를 모두 제거.

set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "🧹 cleaning demo dummy artifacts at $DEMO_DIR"

removed=0
for d in shop blog api bot; do
  if [ -d "$DEMO_DIR/$d" ]; then
    rm -rf "$DEMO_DIR/$d"
    echo "  - removed $d/"
    removed=$((removed + 1))
  fi
done

if [ "$removed" -eq 0 ]; then
  echo "✅ already clean."
else
  echo "✅ removed $removed director(y/ies)."
fi
