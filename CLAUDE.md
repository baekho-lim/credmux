# CLAUDE.md — credmux 0.1

## 프로젝트 개요

**credmux**는 AI 에이전트 시대 바이브코더를 위한 보안 CLI다.
해커톤 당일(2026-04-26) 개발 완료 + 데모 + GitHub 공개가 목표.

**핵심 포지션**: 기존 도구(Gitleaks, TruffleHog)는 git 커밋 전만 막는다.
credmux는 지금 이 순간, 터미널 밖에서도 지킨다.

---

## 아키텍처 (반드시 숙지)

### 두 가지 모드

```
Core Mode    → 모든 터미널, 터미널 없이도 작동
Enhanced Mode → cmux 감지 시 자동 활성화 ($CMUX_SOCKET_PATH 존재 여부)
```

### 3개 컴포넌트

```
SENTINEL  → 플랫폼 해킹 감시 + Telegram/cmux 알림
VAULT     → TruffleHog 기반 스캔 + Keychain 저장 + breach-drill
GUARDIAN  → cmux workspace 격리 + AIM-style watcher
```

### 의존성

```
필수:   trufflehog (brew install trufflehog)
필수:   macOS Keychain (security CLI, 기본 내장)
선택:   cmux ($CMUX_SOCKET_PATH=/tmp/cmux.sock)
선택:   python3 + python-telegram-bot (Telegram bot용)
```

---

## 프로젝트 구조 (이대로 생성할 것)

```
credmux/
├── CLAUDE.md              ← 이 파일
├── README.md
├── go.mod
├── go.sum
├── main.go
├── cmd/
│   ├── root.go
│   ├── scan.go            ← credmux scan
│   ├── breach.go          ← credmux breach-drill [platform]
│   ├── workspace.go       ← credmux workspace open [project]
│   ├── profile.go         ← credmux profile create/list
│   └── bot.go             ← credmux bot start
├── internal/
│   ├── trufflehog/
│   │   └── runner.go      ← TruffleHog subprocess 래퍼
│   ├── keychain/
│   │   └── keychain.go    ← macOS security CLI 래퍼
│   ├── cmux/
│   │   ├── client.go      ← Unix socket JSON-RPC
│   │   ├── detect.go      ← cmux 존재 감지
│   │   ├── notify.go      ← notification.create
│   │   ├── sidebar.go     ← sidebar.set_status, sidebar.log
│   │   └── workspace.go   ← workspace.create, browser.open
│   ├── sentinel/
│   │   ├── monitor.go     ← RSS 감시 루프
│   │   └── platforms.go   ← 플랫폼별 피드 정의
│   ├── watcher/
│   │   └── watcher.go     ← capture-pane 폴링 + 패턴 감지
│   └── config/
│       └── profile.go     ← ~/.credmux/profiles.json
├── bot/
│   ├── bot.py             ← Telegram bot (Python)
│   ├── requirements.txt
│   └── handlers.py
└── demo/
    ├── setup.sh           ← 데모 환경 구성 스크립트
    ├── shop/.env          ← 더미 토큰 파일
    ├── blog/.env.local    ← 더미 토큰 파일
    └── teardown.sh        ← 데모 환경 정리
```

---

## cmux Socket API 스펙 (실제 확인됨)

```
소켓: $CMUX_SOCKET_PATH 또는 /tmp/cmux.sock
프로토콜: JSON-RPC over Unix domain socket
권한: 0600 (소유자 전용)

주요 메서드:
  notification.create  → {"title":"", "subtitle":"", "body":""}
  workspace.create     → {"cwd":""} → {"workspace_id":""}
  workspace.list       → {} → {"workspaces":[...]}
  sidebar.set_status   → {"workspace_id":"", "text":"", "icon":""}
  sidebar.log          → {"level":"info|warning|error", "message":""}
  surface.capture_pane → {"surface_id":""} → {"text":""}
  browser.open         → {"url":""}
  system.ping          → {} → {"pong":true}

환경변수:
  CMUX_SOCKET_PATH   소켓 경로
  CMUX_WORKSPACE_ID  현재 workspace ID
  CMUX_SURFACE_ID    현재 surface ID
```

---

## TruffleHog 스펙 (실제 확인됨)

```bash
# 기본 사용법
trufflehog filesystem ~/projects --json --only-verified --no-update

# exit code
0   → 시크릿 없음
183 → 시크릿 발견 (에러 아님, 정상)
1   → 실행 오류

# JSON 출력 형식 (한 줄씩)
{
  "DetectorName": "Vercel",
  "Raw": "dpl_xxxx",
  "Verified": true,
  "SourceMetadata": {
    "Data": {
      "Filesystem": {
        "file": "/Users/me/projects/shop/.env",
        "line": 3
      }
    }
  }
}
```

---

## macOS Keychain CLI 스펙

```bash
# 저장
security add-generic-password \
  -s "credmux-projectA-vercel" \
  -a "credmux" \
  -w "dpl_xxxx"

# 읽기
security find-generic-password \
  -s "credmux-projectA-vercel" \
  -w

# 삭제
security delete-generic-password \
  -s "credmux-projectA-vercel"
```

---

## 플랫폼별 피드 및 패턴

```go
var Platforms = map[string]Platform{
    "vercel": {
        Name:         "Vercel",
        DetectorName: "Vercel",
        DashboardURL: "https://vercel.com/account/tokens",
        RSS:          "https://www.vercel-status.com/history.rss",
    },
    "github": {
        Name:         "GitHub",
        DetectorName: "GitHub",
        DashboardURL: "https://github.com/settings/tokens",
        RSS:          "https://www.githubstatus.com/history.rss",
    },
    "supabase": {
        Name:         "Supabase",
        DetectorName: "Supabase",
        DashboardURL: "https://app.supabase.com/account/tokens",
        RSS:          "https://status.supabase.com/history.rss",
    },
    "stripe": {
        Name:         "Stripe",
        DetectorName: "Stripe",
        DashboardURL: "https://dashboard.stripe.com/apikeys",
        RSS:          "https://status.stripe.com/history.rss",
    },
    "anthropic": {
        Name:         "Anthropic",
        DetectorName: "Anthropic",
        DashboardURL: "https://console.anthropic.com/account/keys",
        RSS:          "",
    },
}
```

---

## 우선순위 (해커톤 데모 기준)

```
P0 (반드시 작동):
  credmux breach-drill vercel  ← 데모 핵심 1
  Telegram bot scan + 결과     ← 데모 핵심 2
  credmux workspace open       ← 데모 핵심 3

P1 (있으면 좋음):
  credmux scan (standalone)
  cmux sidebar 상태 표시

P2 (해커톤 이후):
  credmux profile create
  토큰 만료 알림
  OAuth 감사
```

---

## 데모 환경

```
demo/ 폴더에 더미 파일이 있음.
실제 토큰처럼 보이지만 패턴만 맞는 가짜 값.
TruffleHog --only-verified 없이 실행하면 탐지됨.
데모 시 HOME 환경변수를 demo/ 로 바꿔서 스캔.
```

---

## 코딩 규칙

```
언어: Go 1.22 (CLI), Python 3.11 (Telegram bot)
에러: 항상 wrapping (fmt.Errorf("context: %w", err))
cmux 없을 때: 반드시 fallback 메시지 출력 (패닉 금지)
시크릿 출력: 절대 평문 출력 금지, 항상 MaskSecret() 사용
테스트: 없어도 됨 (해커톤 시간 부족)
커밋: 영어
```

---

## 빠른 시작 순서

```bash
cd credmux
go mod init github.com/baekho/credmux
go get github.com/spf13/cobra
go get github.com/google/uuid

# TruffleHog 설치 확인
which trufflehog || brew install trufflehog

# 데모 환경 구성
bash demo/setup.sh

# 빌드 테스트
go build -o credmux . && ./credmux --help
```