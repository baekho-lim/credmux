# Claude Code 지시문 — credmux 0.1
## 컨텍스트 엔지니어링 턴바이턴 가이드

**사용 방법**: 아래 지시문을 순서대로 Claude Code에 입력한다.
Turn 1은 컨텍스트 전달, Turn 2부터 병렬 개발 시작.

---

## TURN 1 — 컨텍스트 로딩 (가장 먼저)

```
다음 파일들을 읽고 프로젝트 전체 구조를 파악해라.
읽는 것 외에 아무것도 하지 말고, 파악한 내용을 요약해줘.

1. CLAUDE.md          ← 프로젝트 전체 컨텍스트
2. credmux-TDD-v1.md  ← 기술 설계서 (구현 상세)
3. credmux-PRD-v2.md  ← 제품 요구사항

파악 후 아래 항목을 확인해줘:
- 개발해야 할 P0 기능 3개
- 의존성 설치 필요 여부
- 첫 번째로 만들어야 할 파일
```

---

## TURN 2 — 프로젝트 scaffold + 병렬 계획 수립

```
이제 개발을 시작한다. 아래 두 가지를 동시에 수행해라.

[작업 1] Go 프로젝트 초기화
  go mod init github.com/baekho/credmux
  필요한 패키지:
    github.com/spf13/cobra
    github.com/google/uuid
  
  다음 파일들을 생성해라:
    main.go         (cobra root command, 서브커맨드 등록만)
    cmd/root.go     (root command 정의)
    go.mod / go.sum

[작업 2] 병렬 개발 계획 수립
  다음 3개 컴포넌트를 병렬로 개발할 순서를 계획해라:
  
  Component A: VAULT (breach-drill)
    internal/trufflehog/runner.go
    internal/keychain/keychain.go
    cmd/breach.go
  
  Component B: GUARDIAN (workspace)
    internal/cmux/client.go
    internal/cmux/detect.go
    internal/cmux/notify.go
    internal/cmux/sidebar.go
    internal/cmux/workspace.go
    cmd/workspace.go
  
  Component C: SENTINEL (Telegram bot)
    bot/bot.py
    bot/handlers.py
    bot/requirements.txt

  각 컴포넌트가 독립적으로 작동 가능한지 확인해라.
  (A는 cmux 없어도, B는 trufflehog 없어도 작동해야 함)
```

---

## TURN 3 — Component A 개발: VAULT

```
Component A를 완전히 구현해라. 이것이 데모 핵심 1번이다.

구현 파일:
  internal/trufflehog/runner.go
  internal/keychain/keychain.go
  cmd/breach.go
  cmd/scan.go

핵심 요구사항:
  1. trufflehog filesystem [path] --json --no-update 실행
     --only-verified는 데모에서 제거 (더미 토큰 탐지용)
     exit code 183이 정상 (시크릿 발견)
  
  2. 결과 파싱: DetectorName, Raw(마스킹), file 경로
  
  3. credmux breach-drill vercel 실행 시:
     - 스캔 → vercel 관련 토큰만 필터
     - 결과 출력 (이모지 포함, 보기 좋게)
     - cmux 감지되면 browser.open(dashboardURL) 호출
     - cmux 없으면 URL 텍스트 출력
  
  4. MaskSecret(): 앞 8자리만 표시

데모 환경: demo/ 폴더
  DEMO_HOME=./demo 환경변수로 스캔 경로 지정
  
완료 후 테스트:
  DEMO_HOME=./demo go run . breach-drill vercel
  → shop/.env와 blog/.env.local에서 VERCEL_TOKEN 탐지 확인
```

---

## TURN 4 — Component B 개발: GUARDIAN

```
Component B를 구현해라. 데모 핵심 3번이다.

구현 파일:
  internal/cmux/client.go    ← JSON-RPC over Unix socket
  internal/cmux/detect.go    ← cmux 존재 감지
  internal/cmux/notify.go    ← notification.create
  internal/cmux/sidebar.go   ← sidebar.set_status, sidebar.log
  internal/cmux/workspace.go ← workspace.create, browser.open
  internal/keychain/profile.go ← 프로파일별 토큰 주입
  internal/watcher/watcher.go  ← capture-pane 폴링
  cmd/workspace.go

cmux API 스펙 (CLAUDE.md 참조):
  소켓: /tmp/cmux.sock
  메서드: notification.create, workspace.create,
          sidebar.set_status, sidebar.log,
          surface.capture_pane, browser.open

핵심 요구사항:
  1. cmux 감지: $CMUX_SOCKET_PATH 또는 /tmp/cmux.sock 존재 여부
  
  2. credmux workspace open projectA 실행 시:
     Core Mode (cmux 없음):
       → "eval $(credmux workspace env projectA)" 안내
     Enhanced Mode (cmux 있음):
       → workspace.create 호출
       → sidebar.set_status: "🔐 projectA | tokens isolated"
       → sidebar.log: "credmux: workspace opened"
       → watcher goroutine 시작
  
  3. watcher: 3초마다 capture-pane 폴링
     시크릿 패턴 감지 시:
       → notification.create 빨간 경보
       → sidebar.log error 레벨
  
  4. cmux 없을 때 절대 패닉하지 말 것
     fallback 메시지 출력하고 정상 종료

완료 후 테스트:
  cmux 없을 때: go run . workspace open demo
  → Core Mode fallback 메시지 확인
```

---

## TURN 5 — Component C 개발: SENTINEL (Telegram Bot)

```
Component C를 구현해라. 데모 핵심 2번이다.
Python으로 구현.

파일:
  bot/requirements.txt
  bot/bot.py
  bot/handlers.py

requirements.txt:
  python-telegram-bot==20.7
  python-dotenv==1.0.0
  requests==2.31.0
  feedparser==6.0.10

핵심 요구사항:
  1. /start 명령 → 봇 소개 + 사용법
  
  2. 자동 알림 (SENTINEL):
     Vercel RSS 피드 주기적 체크 (60초마다)
     "security" OR "breach" OR "incident" 키워드 감지 시
     → 인라인 버튼과 함께 알림 메시지 발송
  
  3. 인라인 버튼 UI:
     [✅ 지금 스캔]  [⏰ 나중에]
     
     [✅ 지금 스캔] 클릭 시:
       subprocess로 credmux breach-drill [platform] --json 실행
       결과 파싱 → 포맷팅 → 메시지 발송
       
       발견 없음: "✅ 영향받은 토큰 없음. 안전합니다."
       발견 있음: 파일별 목록 + [🔑 교체 안내] 버튼
     
     [🔑 교체 안내] 클릭 시:
       단계별 교체 방법 텍스트 발송
       "새 토큰을 여기에 붙여넣으세요:" 안내
  
  4. DEMO_HOME 환경변수 지원
     봇이 subprocess 실행 시 DEMO_HOME 전달

  5. 해커톤 데모용 /testbreach 명령:
     실제 RSS 감지 기다리지 않고
     바로 Vercel breach 알림 시뮬레이션

환경변수:
  TELEGRAM_BOT_TOKEN  필수
  DEMO_HOME          선택 (데모 환경 경로)
  CREDMUX_BINARY     선택 (credmux 바이너리 경로, 기본: ./credmux)

완료 후 테스트:
  pip install -r bot/requirements.txt
  TELEGRAM_BOT_TOKEN=xxx DEMO_HOME=./demo python3 bot/bot.py
  → /testbreach 명령으로 데모 플로우 확인
```

---

## TURN 6 — 통합 테스트 + 데모 검증

```
세 컴포넌트를 통합하고 데모 플로우를 완전히 검증해라.

[테스트 1] Core Mode 완전 플로우
  bash demo/setup.sh
  go build -o credmux .
  DEMO_HOME=./demo ./credmux breach-drill vercel
  
  기대 결과:
  🚨 Vercel Security Incident (2026-04-19)
  [스캔 중...]
    demo/shop/.env    VERCEL_TOKEN  ⚠️  ...
    demo/blog/.env.local  VERCEL_TOKEN  ⚠️  ...
  영향: 2개 파일
  → vercel.com/account/tokens

[테스트 2] Enhanced Mode (cmux 있을 때)
  cmux 실행 중인 환경에서:
  DEMO_HOME=./demo ./credmux workspace open demo
  
  기대 결과:
  cmux sidebar에 "🔐 demo | tokens isolated" 표시
  notification 팝오버 발생

[테스트 3] Telegram Bot
  TELEGRAM_BOT_TOKEN=xxx DEMO_HOME=./demo python3 bot/bot.py &
  텔레그램에서 /testbreach 입력
  
  기대 결과:
  봇이 알림 + [✅ 지금 스캔] 버튼 발송
  버튼 클릭 → 스캔 결과 회신

[데모 시나리오 리허설]
  1분 데모 영상 순서대로 3번 연속 실행 확인:
  
  Step 1: cat demo/shop/.env | grep TOKEN (평문 노출)
  Step 2: DEMO_HOME=./demo ./credmux breach-drill vercel (터미널)
  Step 3: 텔레그램 /testbreach → [✅ 지금 스캔] (스마트폰)
  Step 4: cmux workspace open demo (cmux sidebar)

모든 테스트 통과 후 보고해라.
실패한 테스트가 있으면 즉시 수정 후 재테스트.
```

---

## TURN 7 — README + GitHub 공개 준비

```
발표 제출 전 마지막 작업이다.

[작업 1] README.md 생성
  내용:
  - 한 줄 설명 (영어)
  - 문제: "기존 도구는 git만 본다. credmux는 지금 이 순간을 본다."
  - 설치: brew install trufflehog && go install ...
  - 빠른 시작 3개 명령어
  - Core vs Enhanced 표
  - Built at CMUX × AIM Hackathon Seoul 2026

[작업 2] .gitignore 생성
  credmux (바이너리)
  .env*
  demo/shop/
  demo/blog/
  demo/bot/
  __pycache__/
  *.pyc
  bot/.env

[작업 3] go build 최종 확인
  go build -o credmux .
  ./credmux --help
  → 모든 서브커맨드 표시 확인

[작업 4] 최종 데모 스크립트 출력
  demo/run_demo.sh 생성:
  
  #!/bin/bash
  # credmux 데모 실행 스크립트
  echo "=== Step 1: 평문 토큰 노출 ==="
  cat demo/shop/.env | grep TOKEN
  
  echo ""
  echo "=== Step 2: breach-drill ==="
  DEMO_HOME=./demo ./credmux breach-drill vercel
  
  echo ""
  echo "=== Step 3: Telegram bot은 스마트폰 확인 ==="
  echo "→ /testbreach 입력"

완료 후 보고:
  □ go build 성공
  □ README.md 생성 완료
  □ 데모 스크립트 작동 확인
  □ GitHub 공개 준비 완료
```

---

## 병렬 실행 가이드

시간이 없으면 Turn 3, 4, 5를 동시에 진행한다.
Claude Code 세션이 하나라면 아래 순서 추천:

```
우선순위:
  1. Turn 1–2 (scaffold, 5분)
  2. Turn 3 (VAULT, 45분) ← 데모 핵심
  3. Turn 5 (Telegram Bot, 45분) ← 데모 핵심
  4. Turn 4 (GUARDIAN, 30분) ← cmux 연동
  5. Turn 6–7 (통합, 준비, 30분)

총 예상: 2.5시간

Claude Code 세션 여러 개라면:
  Session A → Turn 3 (VAULT)
  Session B → Turn 5 (Telegram Bot)
  Session C → Turn 4 (GUARDIAN)
  → 병렬로 진행 후 Turn 6에서 통합
```

---

## 긴급 상황 대응

```
[TruffleHog 설치 안 됨]
  brew install trufflehog
  설치 안 되면 내장 패턴으로 대체:
    credmux scan 자체 구현 (CLAUDE.md의 패턴 사용)

[cmux 없음]
  Core Mode만 데모
  "cmux Enhanced는 GitHub README에서 확인 가능합니다" 멘트

[Telegram Bot 토큰 없음]
  credmux bot start --mock 옵션 추가
  실제 봇 없이 터미널에서 시뮬레이션

[Go 빌드 실패]
  go mod tidy 실행
  패키지 import 경로 확인
  최악의 경우 bash 스크립트로 breach-drill 구현
```