# Checklist Candidates

> 세션 중 발견된 잠재 이슈/미구현/리스크. 데모/릴리스 직전 일괄 점검용.
> 형식: `- [ ] 항목` — 해소되면 체크 또는 삭제. 새 이슈는 발견 즉시 append.

## 데모 직전 필수 점검

- [ ] **demo/setup.sh / teardown.sh 작성** — PRD에 명시됨. 데모 환경 일괄 구성/정리 스크립트 (현재 .env 더미만 수동 생성 상태)
- [ ] **TruffleHog 첫 스캔 지연** — `--no-update`에도 5–15초 걸릴 수 있음. 데모 영상에서 dead air 위험. 사전 워밍업 또는 스피너 출력 고려
- [ ] **Keychain GUI 권한 프롬프트** — 첫 `security add-generic-password` 호출 시 macOS GUI 동의 팝업 발생 가능. 데모 흐름 끊김 → 사전에 한 번 실행해 권한 캐시
- [ ] **.gitignore 미작성** — `credmux` 바이너리, `demo/shop/`, `demo/blog/`, `__pycache__/`, `bot/.env` 등 제외 필요. public 전 필수

## Component 간 통합 의존

- [ ] **`breach-drill --json` 플래그** — Turn 5(SENTINEL/Telegram)가 subprocess 결과 파싱하려면 필요. 현재 사람-친화 텍스트 출력만
- [x] ~~**`cmux.OpenBrowser` stub 메시지 톤**~~ — Component B 본 구현 완료, stderr fallback 메시지로 교체됨 (2026-04-26)
- [x] ~~**`breach-drill --json` 플래그**~~ — Component C 작업 시 추가 완료. unknown platform / 빈 findings / TruffleHog 미설치 모두 valid JSON (2026-04-26)
- [x] ~~**demo/setup.sh 작성**~~ — 5개 플랫폼 더미 생성 + setup.sh 자체가 detector에 안 잡히도록 prefix/suffix 분할 (2026-04-26)
- [ ] **cmux 실 환경 통합 테스트 부재** — `workspace open` Enhanced 경로(workspace.create, sidebar.set_status, watcher)는 cmux 소켓 없는 현재 환경에선 검증 불가. 데모 직전 cmux 떠 있는 머신에서 1회 smoke-test 필수 (Turn 6 테스트 2 partial: Core fallback만 ✓)
- [ ] **Telegram 실 가동 검증** — Turn 6 테스트 3 partial: deps install/import/boot/handler 데이터 경로 ✓, BotFather 토큰으로 실제 송수신 미검증. 데모 전 봇 한 번 띄워서 /testbreach 응답 확인 필요
- [ ] **Python 3.13+ feedparser 호환** — feedparser 6.0.10이 stdlib `cgi`(3.13 제거됨) import. `legacy-cgi==2.6.1; python_version >= "3.13"` 추가로 우회. 향후 feedparser가 cgi 의존 제거하면 polyfill 라인 삭제
- [ ] **watcher의 동일-hit 억제 로직** — `lastHit` 단일 변수로 비교. 서로 다른 토큰이 같은 pane에 같이 나오면 첫 hit 이후 둘째 토큰 알람이 누락될 수 있음. set/map 기반으로 업그레이드 후보

## 미구현 fallback

- [ ] **TruffleHog 미설치 시 내장 패턴 fallback** — CLAUDE.md 긴급 대응에 명시. 현재는 `ErrNotInstalled`로 종료만. 데모 직전 trufflehog 못 설치 시 백업 경로 없음
- [ ] **`uuid` 의존성 자동 제거됨** — `go mod tidy`가 require에서 빼버림. Profile/Workspace 본격 구현 시 다시 import 필요
- [ ] **Stripe/Anthropic detector 더미 미매칭** — TruffleHog 3.95.2에서 `sk_live_` 99자, `sk-ant-api03-` 95자+AA suffix 모두 시도해도 0 매칭. detector regex 정확한 형식 + keyword 컨텍스트 재조사 필요. 데모는 Vercel/GitHub만 사용 가능
- [ ] **Supabase detector 부재** — TruffleHog 3.95.2에 Supabase detector 자체 없음 (`--include-detectors=Supabase` 시 "unrecognized detector type" 에러). breach-drill의 platforms map에서 supabase는 영영 0 finding. 자체 패턴 구현 또는 platforms map에서 supabase 제거 검토

## 운영 품질

- [ ] **`Verified=false` 노이즈** — 데모 더미는 unverified로 정상. 운영 환경에서 unverified가 폭증하면 시그널/노이즈 문제. `--only-verified` 토글 옵션 고려
- [ ] **`profile.go`의 defaultProfile 동일 매핑** — 어떤 프로젝트 이름이든 vercel/github/supabase/stripe/anthropic 5개 platform→env 매핑 반환. `~/.credmux/profiles.json` 도입 시 이 fallback이 사용자 의도와 충돌할 수 있음 → 명시적 profile 부재 시 빈 매핑 또는 경고 출력 검토
- [ ] **`MaskSecret` 노출량** — 24자 토큰의 첫 8자 노출(33%). 운영에선 4자 또는 해시 prefix가 더 안전할 수도
- [ ] **TruffleHog 자체 에러 시 종료 코드** — 현재 stderr만 찍고 진행. CI에서는 명확한 non-zero exit가 필요할 수도

## 문서 동기화

- [ ] **README.md** — 현재 scaffold 상태. Component A 동작/사용법 반영 필요. Turn 7에 명시됨
- [ ] **SESSION-LOG.md** — Component A 작업 내역 기록 필요. session end protocol 발동 시
