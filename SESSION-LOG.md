# SESSION-LOG — cremux

> Last Session always at top.

## Last Session — 2026-04-26 (일) Hackathon Day

VAULT / GUARDIAN / SENTINEL 세 컴포넌트 본체 1차 통과.

**구현**
- `cmd/`: root, scan, breach-drill (`--json`), workspace (open/env), profile, bot
- `internal/trufflehog`: subprocess runner, `MaskSecret`, exit-183 정상 처리
- `internal/keychain`: macOS `security` 래퍼 + Profile loader
- `internal/cmux`: JSON-RPC client + notify/sidebar/workspace/browser
- `internal/watcher`: capture-pane 폴링 (8개 토큰 패턴)
- `bot/`: `python-telegram-bot` 20.7 기반 SENTINEL — `/start`, `/testbreach`, 인라인 키보드, 60s RSS poll
- `demo/setup.sh`: 5개 플랫폼 더미 토큰 (prefix/suffix 분할)

**검증 통과**
- `DEMO_HOME=./demo ./credmux breach-drill vercel` → 2건 매칭, 마스킹, URL 안내
- `--json` 모드 모든 에러 경로 valid JSON
- `./credmux workspace open demo` Core fallback 정상
- TruffleHog 매칭: Vercel ✅, GitHub ✅
- Bot subprocess 시뮬레이션 (Python `subprocess` + `json.loads`) — 한국어 렌더까지 통과

**Next**
- `CHECKLIST-CANDIDATES.md` 미해결 항목 (Stripe/Anthropic detector 패턴, Supabase 부재, demo/teardown.sh, README 본문, cmux 실 환경 smoke-test)
- README + Turn 6 통합 리허설

---

## 2026-04-26 (일) — Scaffold

- 빈 레포 생성 (Claude). bhOS 컨벤션 따른 디렉토리.

---
