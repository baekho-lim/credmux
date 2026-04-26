# credmux PRD v2.0
## Product Requirements Document

**작성일**: 2026-04-26  
**버전**: 2.0 (기술 리서치 + 컨설팅 반영)  
**해커톤**: CMUX × AIM Intelligence Hackathon Seoul  
**작성자**: BH Lim (baekho.io)

---

## 1. 제품 개요

### 1.1 한 줄 정의

> **credmux** — AI 에이전트 시대의 보안 수호자. 모든 터미널에서 작동하며, cmux에서 네이티브 시너지가 나고, 비개발자는 Telegram으로 쓴다.

### 1.2 네이밍

`cred`ential + c`mux` — 자격증명 관리와 cmux 에이전트 터미널 철학의 결합.

### 1.3 포지셔닝

```
기존 도구들의 세계:
  Gitleaks / TruffleHog  → git + CI/CD 전용, 개발자 전용
  API Stronghold         → OpenClaw 전용, 클라우드 서비스, 비오픈소스
  GitHub Secret Scanning → GitHub 레포 전용
  OpenClaw security audit → 사후 탐지만, 자동화 없음

credmux의 세계:
  로컬 + 오픈소스 + 모든 터미널 + 모든 에이전트
  + 비개발자 Telegram 인터페이스
  + cmux 네이티브 Enhanced 모드
  + 사전 예방 + 사후 대응
```

---

## 2. 배경 및 문제 정의

### 2.1 시장 데이터

- **63%** 바이브코딩 사용자가 비개발자 (Stanford 연구, 2026)
- **91.5%** 바이브코딩 앱에서 보안 취약점 발견 (Q1 2026)
- **60%+** 생산 앱이 API 키를 공개 레포에 노출
- **28.65M** 개 시크릿이 2025년 공개 GitHub에 유출 (34% 증가)
- **AI 코드**가 인간 코드 대비 **2.74배** 높은 취약점 생성률
- OpenClaw **ClawHavoc** 공격: 800+ 악성 스킬, 수만 명 API 키 탈취
- **CVE-2026-25253** (CVSS 8.8): OpenClaw 원클릭 RCE
- **Issue #9627**: OpenClaw update 시 API 키 평문 노출 버그
- Vercel 해킹 (2026-04-19): Context.ai OAuth → 환경변수 전체 노출

### 2.2 핵심 문제

**문제 1 (개발자/터미널 사용자):**
> "어느 에이전트가 어느 토큰으로 무엇을 하고 있는지 모른다"

**문제 2 (비개발자 바이브코더):**
> "Vercel이 뚫렸다는 뉴스를 봤는데 내가 지금 뭘 해야 하는지 모른다"

**문제 3 (OpenClaw/에이전트 생태계):**
> "모든 스킬이 모든 API 키에 접근할 수 있다. 악성 스킬 하나면 전부 털린다"

---

## 3. 타깃 사용자

### 3.1 Primary Persona — 터미널 바이브코더

```
누구: Claude Code, cmux, OpenClaw, Hermes 등 CLI 에이전트를 쓰는 개발자
특징:
  - 프로젝트 5개+ 병렬, .env 수십 개
  - AI 에이전트 3개+ 동시 실행
  - 보안 알지만 설정할 시간 없음
  - 멀티머신 (맥북 + 맥미니) 운영

Pain:
  "프로젝트 A 에이전트가 프로젝트 B 토큰 읽어서 커밋했다"
  "Vercel 해킹 났는데 내 프로젝트 12개 중 어디가 영향받는지 모른다"

cmux 사용 시 추가 Pain:
  "10개 에이전트 병렬 중 어느 게 시크릿 건드리는지 sidebar에 안 보인다"
```

### 3.2 Secondary Persona — 비개발자 바이브코더

```
누구: Claude Cowork, Lovable, Bolt.new로 앱 만드는 비개발자
특징:
  - 카카오톡 메모 / 노션에 토큰 저장
  - CLI 모름, 터미널 안 씀
  - 보안 위협 자체를 인식 못함
  - Telegram/Discord 봇 운영 중

Pain:
  "Vercel 털렸다는데 나는 어떻게 해야 해요?"
  "BotFather 토큰 어디 저장해야 하는지 모르겠어요"
```

### 3.3 Tertiary Persona — 에이전트 공유/임대자

```
누구: OpenClaw 스킬 개발자, 에이전트 커뮤니티 운영자
특징:
  - 에이전트를 커뮤니티에 빌려줌
  - Rate limit, 만료일 설정 필요
  - $80 과금 사고 직접 경험

Pain:
  "에이전트 빌려줄 때 내 API 키를 통째로 줄 수밖에 없다"
```

---

## 4. 제품 설계

### 4.1 핵심 철학

```
credmux는 두 가지 모드로 작동한다:

Core Mode (모든 사용자)
  어떤 터미널이든, 터미널 없이도 작동
  scan / breach-drill / Telegram bot
  의존성: TruffleHog, macOS Keychain, python-telegram-bot

Enhanced Mode (cmux 감지 시 자동 활성화)
  cmux socket API를 통해 추가 기능 언락
  workspace 격리 / sidebar 상태 / 실시간 경보 / browser 연동
  의존성: cmux $CMUX_SOCKET_PATH

"cmux가 없어도 작동한다. cmux가 있으면 더 강력해진다."
```

### 4.2 3개 컴포넌트

#### SENTINEL — Breach Notifier

외부 플랫폼 해킹 시 사용자보다 먼저 알고 행동한다.

```
작동 방식:
  RSS/API로 보안 인시던트 상시 감시
  대상: Vercel, Supabase, GitHub, Stripe, Cloudflare,
        Anthropic, OpenAI, Railway, PlanetScale, Slack, npm

  인시던트 감지 → 내 Telegram/Discord/Slack 봇으로 알림
  → 버튼 클릭 → VAULT scan 자동 실행 → 결과 회신

cmux Enhanced:
  인시던트 감지 → cmux notification.create (빨간 경보)
  → cmux sidebar.log ("🚨 Vercel breach detected")
  → 사용자 응답 시 breach-drill 자동 실행
```

#### VAULT — Discovery + Rotation

내 맥 전체에 흩어진 시크릿을 찾고, 안전하게 저장하고, 사고 시 교체한다.

```
Discovery (TruffleHog subprocess):
  trufflehog filesystem ~ --json --only-verified
  → 700+ 패턴으로 유효한 시크릿만 탐지
  → 결과 파싱 + 위험도 분류 + 사용자 친화 포맷

breach-drill [platform]:
  영향 프로젝트 스캔 → 교체 필요 토큰 목록
  → cmux browser [dashboard URL] (cmux Enhanced)
  → 또는 안내 URL 출력 (Core Mode)
  → 새 토큰 입력 → Keychain 저장

Keychain 저장:
  security add-generic-password -s credmux-[project] -a [service] -w [token]
  → .env 파일 없음, 디스크에 평문 없음
```

#### GUARDIAN — Agent Isolation

에이전트가 접근할 수 있는 자격증명 범위를 프로젝트 단위로 격리한다.

```
workspace open [project]:
  macOS Keychain에서 [project] 토큰만 로드
  환경변수로 주입 (op run 방식)
  다른 프로젝트 토큰 접근 불가

  Core Mode: 새 shell session에 격리 주입
  cmux Enhanced:
    workspace.create → 새 cmux workspace 생성
    sidebar.set_status: "🔐 [project] | tokens isolated"
    capture-pane 폴링 → 시크릿 패턴 감지
    → notification.create 빨간 경보

git identity 자동 전환:
  git config user.email → [project]@credmux.local
  SSH key context 전환
```

---

## 5. 사용자 시나리오

### 5.1 시나리오 A — 터미널 사용자 (Core Mode)

```
상황: Vercel 해킹 뉴스를 트위터에서 봤다

$ credmux breach-drill vercel

🚨 Vercel Security Incident (2026-04-19)
   벡터: Context.ai OAuth → 환경변수 노출

[스캔 중... TruffleHog 700+ 패턴 적용]
  ~/projects/shop/.env     VERCEL_TOKEN  ✗ verified active
  ~/projects/blog/.env     VERCEL_TOKEN  ✗ verified active
  ~/.zshrc                 GITHUB_TOKEN  ✗ verified active

영향: 2개 프로젝트, 3개 유효 토큰

[교체 안내]
  1. vercel.com/account/tokens 에서 재발급
  2. 완료 후 새 토큰 입력: _____
  → Keychain에 저장 완료 ✓

⏱  47초
```

### 5.2 시나리오 B — cmux 사용자 (Enhanced Mode)

```
상황: cmux에서 10개 에이전트 병렬 실행 중
      projectA 에이전트가 STRIPE_SECRET 출력 시도

cmux sidebar:
  🟢 proj-A  claude 실행 중
  🟢 proj-B  codex 실행 중
  ...

  → 갑자기:
  🔴 proj-A  ⚠️ SECRET DETECTED

cmux 알림 팝오버:
  "STRIPE_SECRET 노출 감지
   projectA workspace에서 차단됨"

$ credmux workspace open projectA
  → 이미 격리됨. 다른 프로젝트 토큰 접근 불가 상태
  → cmux sidebar: "🔐 projectA | STRIPE only | watching"
```

### 5.3 시나리오 C — 비개발자 Telegram

```
상황: Telegram에 credmux 봇 연결됨

[봇 메시지]
🤖 credmux: 🚨 보안 경고
Vercel 해킹 사고 감지 (2026-04-19)
내 프로젝트 영향 확인할까요?

[✅ 스캔하기] [⏰ 나중에]

→ [✅ 스캔하기] 클릭

🤖 credmux: 스캔 중...
~/projects/shop/.env  VERCEL_TOKEN ⚠️
~/projects/blog/.env  VERCEL_TOKEN ⚠️
2개 프로젝트 영향

어떻게 할까요?
[🔑 교체 방법 안내] [📋 나중에 볼게요]

→ [🔑 교체 방법 안내] 클릭

🤖 credmux:
1단계. vercel.com/account/tokens 접속
2단계. 기존 토큰 삭제 → 새 토큰 발급
3단계. 새 토큰을 여기에 붙여넣으세요:

[사용자가 토큰 붙여넣기]

✅ 안전하게 저장됐습니다!
   다음부터 이 토큰은 암호화된 Keychain에서 관리됩니다.
```

---

## 6. 기능 범위 (MVP — 해커톤)

### 6.1 Must Have (오늘 작동)

| 기능 | 설명 | 모드 |
|------|------|------|
| `credmux breach-drill [platform]` | TruffleHog 스캔 + 교체 안내 | Core |
| Telegram Breach Bot | 인시던트 감지 + 버튼 UI + 결과 회신 | Core |
| `credmux workspace open [project]` | Keychain 격리 주입 + cmux sidebar 상태 | Enhanced |
| AIM-style watcher | capture-pane 폴링 + 시크릿 감지 + notify | Enhanced |

### 6.2 Nice to Have (시간 남으면)

- `credmux scan` standalone 명령어
- 토큰 만료 경고
- git identity 자동 전환

### 6.3 Out of Scope (해커톤 이후)

- Discord/Slack 봇
- ClawHub 스킬 배포
- 감사 로그 PDF
- 팀 권한 대시보드
- 한국어 마법사 UI

---

## 7. 가격 모델

| 플랜 | 가격 | 핵심 기능 |
|------|------|---------|
| **Free** | $0 | Core 전부, workspace 2개, Telegram bot |
| **Pro** | $9/월 | 해킹 알림 자동화, 토큰 만료 알림, OAuth 감사, workspace 무제한 |
| **Agent** | $12/월 | 봇 헬스 대시보드, 멀티머신 환경 분리, 임대 프로파일 |
| **Full** | $19/월 | Pro + Agent 전부 |
| **Team** | $29/인/월 | 감사 로그 PDF, 팀 권한, 컴플라이언스 리포트 |

---

## 8. 성공 지표

### 해커톤 (오늘)
- [ ] 대상 수상 (₩3,000,000)
- [ ] Austin Wang: "cmux를 가장 잘 이해한 팀"
- [ ] AIM: "AI 보안 위협을 정확히 다룬 팀"
- [ ] GitHub 레포 public 전환

### D+30
- GitHub Stars 500+
- brew install 100+
- cmux GitHub RFC Issue 오픈

### D+90
- Stars 1,500+
- Pro 전환 50명 ($450 MRR)
- cmux PR 머지 1개

---

## 9. 피칭 포인트 (5분)

```
[00:00] "OpenClaw 쓰시는 분 계세요?
         마이크로소프트가 뭐라 했는지 아세요?
         '표준 워크스테이션에서 실행 부적합'.
         그래도 다들 씁니다. 저도요. 그래서 만들었습니다."

[00:30] [Telegram 데모 — 스마트폰 화면]
        봇 알림 → 스캔 → 결과 → 교체 안내. 30초.

[01:30] [cmux 데모]
        workspace open → sidebar 격리 표시
        에이전트 시크릿 시도 → 즉시 경보. cmux 없으면 못 함.

[03:00] "Gitleaks와 TruffleHog은 git만 봅니다.
         에이전트 런타임을 모릅니다.
         API Stronghold는 OpenClaw 전용 클라우드 서비스입니다.
         credmux는 로컬, 오픈소스, 모든 터미널,
         cmux에서 네이티브, 비개발자는 Telegram으로."

[04:30] "오늘 런칭합니다. github.com/baekho/credmux"
```

---

## Appendix. 용어 정리

| 용어 | 정의 |
|------|------|
| Core Mode | 모든 터미널에서 작동하는 기본 레이어 |
| Enhanced Mode | cmux 감지 시 자동 활성화되는 추가 레이어 |
| SENTINEL | 외부 플랫폼 해킹 감시 + 알림 컴포넌트 |
| VAULT | 시크릿 탐지 + Keychain 저장 + 교체 컴포넌트 |
| GUARDIAN | 에이전트 자격증명 격리 컴포넌트 |
| TruffleHog | 700+ 패턴 + 라이브 API 검증 오픈소스 스캐너 |
| ClawHavoc | OpenClaw ClawHub 공급망 공격 (800+ 악성 스킬) |
| Issue #9627 | OpenClaw update 시 API 키 평문 노출 버그 |
| breach-drill | 플랫폼 해킹 시 즉각 대응 명령어 |
