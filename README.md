# credmux

> Security guardian for vibe-coders in the AI agent era.
> Works in any terminal, gains superpowers in cmux.

## The problem

기존 도구(Gitleaks, TruffleHog)는 **git 커밋 직전만** 본다. 그 사이의 모든 일 — 에이전트가 토큰을 읽고, 평문이 `.env`에 떨어지고, 외부 플랫폼이 해킹되는 순간 — 은 무방비다.

**credmux는 지금 이 순간을 본다.** 모든 터미널에서, cmux 안에서, 비개발자의 텔레그램에서.

## What it does

`credmux`는 3개 컴포넌트를 한 바이너리에 묶는다.

| 컴포넌트 | 역할 |
|----------|------|
| **VAULT** | TruffleHog 700+ 패턴 + macOS Keychain. `breach-drill <platform>`으로 영향받은 토큰을 찾고 교체 페이지로 안내 |
| **GUARDIAN** | cmux workspace 격리 + AIM 스타일 watcher. 에이전트가 시크릿을 만지면 사이드바·알림으로 경보 |
| **SENTINEL** | Vercel 등 RSS 감시 + 텔레그램 봇. CLI 모르는 사람도 사고 발생 시 봇 메시지 한 번이면 영향 확인 |

## Two modes

| | Core Mode | Enhanced Mode |
|---|---|---|
| 트리거 | 항상 | `$CMUX_SOCKET_PATH` 감지 시 자동 |
| 의존성 | `trufflehog`, macOS Keychain | + cmux 데몬 |
| `breach-drill` | 텍스트 결과 + 교체 URL | + cmux 브라우저 자동 오픈 |
| `workspace open` | `eval $(...)` 안내문 | + 사이드바 상태 + 시크릿 watcher |
| Telegram bot | subprocess로 동일 동작 | (동일) |

> *cmux가 없어도 작동한다. cmux가 있으면 더 강력해진다.*

## Install

```bash
brew install trufflehog                                        # required (700+ detectors)

# Option A — go install (single binary, 가장 간단)
go install github.com/baekho-lim/credmux@latest

# Option B — clone + build (소스/데모 같이 받기)
git clone https://github.com/baekho-lim/credmux.git
cd credmux && go build -o credmux . && ./credmux --help
```

## Quick start

```bash
# 1. 데모 환경 만들기 — 3 위치에 가짜 토큰 분산
bash demo/setup.sh

# 2. Vercel 사고 났다고 가정 — 영향 토큰 즉시 탐지 + 마스킹 출력
DEMO_HOME=./demo ./credmux breach-drill vercel

# 3. (선택) 텔레그램 봇 가동 — 스마트폰에서 /testbreach
python3 -m venv .venv-bot && source .venv-bot/bin/activate
pip install -r bot/requirements.txt
TELEGRAM_BOT_TOKEN=<your-bot-token> DEMO_HOME=./demo python3 bot/bot.py
```

전체 데모 플로우 한 줄 — 단계별 [Y/n] 프롬프트로 페이스 조절:

```bash
./credmux demo                # 3-step 인터랙티브 (1분)
./credmux demo --auto         # 프롬프트 없이 자동 실행 (CI / 에이전트)
```

각 단계마다 한 줄 narrative + 명령 출력. Enter만 눌러도 진행.

### Run from any AI agent

Codex / Kimi / Claude Code 같은 에이전트 터미널에서 `./credmux demo --auto` 실행 후 **결과를 자동 컨설팅** 받기 — [`demo/AGENT-PROMPT.md`](./demo/AGENT-PROMPT.md)의 한국어/영문 프롬프트를 에이전트에 그대로 paste.

## CLI

| Command | Description |
|---------|-------------|
| `credmux scan [path]` | TruffleHog로 경로 스캔 (마스킹 출력) |
| `credmux breach-drill <platform>` | 플랫폼별 토큰 필터 + 교체 안내. `--json`으로 머신-친화 |
| `credmux workspace open <project>` | Core: env injection 안내 / Enhanced: cmux 격리 워크스페이스 |
| `credmux workspace env <project>` | `eval $(...)` 용 export 라인 출력 |
| `credmux bot start` | 텔레그램 봇 안내 (실 가동은 `python3 bot/bot.py`) |
| `credmux demo` | 번들 해커톤 데모 실행 (`demo/run_demo.sh`) |

지원 플랫폼: `vercel`, `github`, `supabase`, `stripe`, `anthropic`
(현재 데모로 매칭 검증된 detector: Vercel, GitHub. Stripe/Anthropic/Supabase는 [CHECKLIST-CANDIDATES.md](./CHECKLIST-CANDIDATES.md) 참조)

## Stack

Go 1.22 (CLI) · Python 3.11+ (Telegram bot, 3.13+ 호환) · macOS Keychain (`security` CLI) · cmux JSON-RPC over Unix socket · TruffleHog 3.95+ · python-telegram-bot 20.7

## Status

해커톤 1일차 — 모든 P0 기능 작동. 미해결/한계는 [CHECKLIST-CANDIDATES.md](./CHECKLIST-CANDIDATES.md).

## License

TBD (해커톤 직후 결정).

---

Built at **CMUX × AIM Intelligence Hackathon Seoul 2026** · 2026-04-26
