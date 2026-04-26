# Agent Demo Prompt

> credmux를 AI 에이전트(Codex / Kimi / Claude Code) 안에서 실행하고
> 결과를 자동 분석·컨설팅하도록 유도하는 표준 프롬프트.

## 사용 방법

1. 에이전트 터미널 세션에서 `cd ~/.../credmux` (저장소 루트)
2. 아래 프롬프트를 그대로 paste

---

## Prompt — 한국어 (발표용 권장)

```
너는 AI 보안 컨설턴트다. 다음 절차를 수행해.

1. 터미널에서 다음 두 명령을 실행하고 출력 전체를 캡처해라:
   bash demo/setup.sh
   ./credmux demo

2. 또한 머신-친화 데이터를 얻기 위해 다음을 실행해라:
   DEMO_HOME=./demo ./credmux breach-drill vercel --json

3. 출력을 분석해 다음을 한국어로 보고해라:
   - 노출된 토큰 개수와 파일 경로 (마스킹된 prefix만 인용)
   - 각 토큰의 위험도 (verified / unverified, 같은 토큰 재사용 여부)
   - 영향받은 프로젝트 (파일 경로에서 추정)
   - 즉시 수행할 액션 (rotate URL, 우선순위 1·2·3)
   - 같은 사고가 다음에 안 일어나게 할 운영 권고 1줄

4. 마지막으로 사용자에게 다음을 제안해라:
   - "Telegram bot으로 비개발자 동료에게도 알릴까?"
   - "cmux workspace로 격리해서 다음 에이전트 작업 시작할까?"

출력은 5줄 이내 핵심 + 액션 표 형태로.
```

---

## Prompt — 영문 (외국 심사관 / OSS 노출용)

```
You are an AI security consultant. Do this:

1. In the terminal, run and capture full output:
   bash demo/setup.sh
   ./credmux demo

2. Also fetch the machine-readable payload:
   DEMO_HOME=./demo ./credmux breach-drill vercel --json

3. Analyze and report (in English) the following:
   - Number of exposed tokens and their file paths (cite masked prefixes only)
   - Severity per token (verified vs unverified, reuse across files)
   - Affected projects inferred from paths
   - Immediate actions with priority (rotate URL, ordered 1·2·3)
   - One-liner operational recommendation to prevent recurrence

4. Then ask the user:
   - "Notify a non-dev teammate via the Telegram bot?"
   - "Open a cmux workspace to isolate the next agent run?"

Output ≤5 bullets + a tight action table.
```

---

## 데모 흐름 (발표 시연용)

| Step | 누가 | 무엇 |
|------|------|------|
| 1 | 발표자 | 에이전트(Codex/Kimi/Claude Code)에 위 한국어 프롬프트 paste |
| 2 | 에이전트 | `bash demo/setup.sh`, `./credmux demo`, `breach-drill --json` 자동 실행 |
| 3 | 에이전트 | 출력 분석 → 한국어 컨설팅 메시지 + 액션 표 |
| 4 | 발표자 | 스마트폰에서 Telegram `/testbreach` → 비개발자 시연 |
| 5 | 발표자 | cmux 환경에서 `./credmux workspace open demo` → 사이드바 격리 시연 |

**핵심 메시지**: credmux는 데이터 소스. 에이전트는 그걸 읽고 자연어로 컨설팅. 사람은 결정만 한다.
