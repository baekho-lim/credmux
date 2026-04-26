package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/baekho-lim/credmux/internal/trufflehog"
	"github.com/spf13/cobra"
)

var (
	auditPlatform  string
	auditTarget    string
	auditNoOpen    bool
	auditReportDir string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit local credentials and emit a multi-step Markdown report",
	Long: `audit는 현재 환경(또는 --target)을 자동 점검하고, 단계별로 분리된
마크다운 보고서를 생성한다.

생성물 — ~/.credmux/reports/<timestamp>/ 안에:
  00-summary.md          (요약 + 다음 단계 링크)
  01-exposure.md         (TruffleHog 노출 진단)
  02-breach-response.md  (--platform 사고 대응 시뮬)
  03-next-actions.md     (지속 보호 — Telegram / cmux 가이드)

마지막에 file:// URL을 출력하고 (--no-open이 아니면) 기본 앱으로 자동
오픈한다. 발표/컨설팅에서 클릭 한 번으로 결과 시연 가능.

  --platform <name>   사고 대응 시뮬 대상 (vercel|github|stripe|anthropic|supabase)
  --target <path>     스캔 경로 (default $DEMO_HOME or $HOME)
  --report-dir <path> 보고서 저장 위치 명시
  --no-open           자동 오픈 끄기`,
	RunE: runAudit,
}

func init() {
	auditCmd.Flags().StringVar(&auditPlatform, "platform", "vercel",
		"platform to simulate breach response for")
	auditCmd.Flags().StringVar(&auditTarget, "target", "",
		"scan target dir (default: $DEMO_HOME, then $HOME)")
	auditCmd.Flags().BoolVar(&auditNoOpen, "no-open", false,
		"do not auto-open the summary in the default app")
	auditCmd.Flags().StringVar(&auditReportDir, "report-dir", "",
		"explicit report directory (default: ~/.credmux/reports/<timestamp>)")
}

func runAudit(_ *cobra.Command, _ []string) error {
	plat, ok := platforms[strings.ToLower(auditPlatform)]
	if !ok {
		return fmt.Errorf("unknown --platform %q (vercel|github|supabase|stripe|anthropic)", auditPlatform)
	}

	target := auditTarget
	if target == "" {
		target = scanRoot()
	}

	reportDir, err := resolveReportDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(reportDir, 0700); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	startedAt := time.Now()

	fmt.Println("🔐 credmux audit")
	fmt.Printf("   target:   %s\n", target)
	fmt.Printf("   platform: %s\n", plat.Name)
	fmt.Printf("   report:   %s\n\n", reportDir)

	// Step 1 — Exposure scan (single TruffleHog run, shared across steps)
	fmt.Print("▶ Step 1/3 — 노출 진단 ... ")
	findings, scanErr := trufflehog.Scan(target)
	if scanErr != nil {
		fmt.Println("scan error:", scanErr)
		// 계속 진행해서 보고서에 에러를 기록한다.
	}
	step1Path := filepath.Join(reportDir, "01-exposure.md")
	if err := writeStep1(step1Path, target, findings, scanErr); err != nil {
		return err
	}
	fmt.Printf("done (%d finding%s)\n", len(findings), pluralS(len(findings)))
	fmt.Printf("  → file://%s\n\n", step1Path)

	// Step 2 — Platform-specific breach response simulation
	fmt.Printf("▶ Step 2/3 — %s 사고 대응 시뮬 ... ", plat.Name)
	matched := filterByDetector(findings, plat.DetectorName)
	step2Path := filepath.Join(reportDir, "02-breach-response.md")
	if err := writeStep2(step2Path, plat, matched); err != nil {
		return err
	}
	fmt.Printf("done (%d affected)\n", len(matched))
	fmt.Printf("  → file://%s\n\n", step2Path)

	// Step 3 — Next actions guide
	fmt.Print("▶ Step 3/3 — 지속 보호 ... ")
	step3Path := filepath.Join(reportDir, "03-next-actions.md")
	if err := writeStep3(step3Path, plat, target); err != nil {
		return err
	}
	fmt.Println("done")
	fmt.Printf("  → file://%s\n\n", step3Path)

	// Summary
	summaryPath := filepath.Join(reportDir, "00-summary.md")
	if err := writeSummary(summaryPath, target, plat, findings, matched, startedAt); err != nil {
		return err
	}

	fmt.Println("──────────────────────────────────────────────")
	fmt.Println("📄 Audit complete.")
	fmt.Printf("   Open: file://%s\n", summaryPath)

	if !auditNoOpen {
		if err := openInDefaultApp(summaryPath); err != nil {
			fmt.Fprintf(os.Stderr, "   (auto-open failed: %v — open the URL above manually)\n", err)
		}
	}
	return nil
}

func resolveReportDir() (string, error) {
	if auditReportDir != "" {
		return auditReportDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	ts := time.Now().Format("20060102-150405")
	return filepath.Join(home, ".credmux", "reports", ts), nil
}

func openInDefaultApp(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Run()
	case "linux":
		return exec.Command("xdg-open", path).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Run()
	}
	return fmt.Errorf("unsupported platform %s", runtime.GOOS)
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// ---------------------------------------------------------------------------
// Markdown writers
// ---------------------------------------------------------------------------

func writeStep1(path, target string, findings []trufflehog.Finding, scanErr error) error {
	var b strings.Builder
	fmt.Fprintln(&b, "# Step 1 — 노출 진단")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "> `%s`에서 평문으로 떨어진 자격증명을 TruffleHog 700+ 패턴으로 스캔.\n", target)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Command")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```bash")
	fmt.Fprintf(&b, "credmux scan %s\n", target)
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b)

	if scanErr != nil {
		fmt.Fprintf(&b, "## ⚠️ Scan error\n\n```\n%v\n```\n\n", scanErr)
	}

	fmt.Fprintf(&b, "## Findings — %d total\n\n", len(findings))
	if len(findings) == 0 {
		fmt.Fprintln(&b, "✅ 노출된 토큰 없음. 안전한 상태.")
	} else {
		fmt.Fprintln(&b, "| Detector | File | Line | Masked |")
		fmt.Fprintln(&b, "|----------|------|------|--------|")
		for _, fnd := range findings {
			fmt.Fprintf(&b, "| %s | `%s` | %d | `%s` |\n",
				fnd.DetectorName, fnd.File, fnd.Line, trufflehog.MaskSecret(fnd.Raw))
		}
		fmt.Fprintln(&b)

		byDet := groupByDetector(findings)
		fmt.Fprintln(&b, "### By detector")
		fmt.Fprintln(&b)
		for _, det := range sortedKeys(byDet) {
			fmt.Fprintf(&b, "- **%s** — %d 건\n", det, len(byDet[det]))
		}
		fmt.Fprintln(&b)

		files := uniqueFileList(findings)
		fmt.Fprintf(&b, "### Affected files — %d\n\n", len(files))
		for _, p := range files {
			fmt.Fprintf(&b, "- `%s`\n", p)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "---")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "[← Summary](./00-summary.md) · [Next → Step 2](./02-breach-response.md)")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeStep2(path string, plat platform, matched []trufflehog.Finding) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Step 2 — %s 사고 대응 시뮬\n\n", plat.Name)
	fmt.Fprintf(&b, "> %s가 뚫렸다고 가정. 영향받은 토큰을 즉시 찾아 교체 페이지로 안내.\n\n", plat.Name)

	fmt.Fprintln(&b, "## Command")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```bash")
	fmt.Fprintf(&b, "credmux breach-drill %s\n", strings.ToLower(plat.Name))
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "## Impact — %d %s token%s\n\n", len(matched), plat.Name, pluralS(len(matched)))
	if len(matched) == 0 {
		fmt.Fprintln(&b, "✅ 영향받은 토큰 없음. 안전합니다.")
	} else {
		fmt.Fprintln(&b, "| File | Line | Masked | Status |")
		fmt.Fprintln(&b, "|------|------|--------|--------|")
		for _, fnd := range matched {
			status := "⚠️ unverified"
			if fnd.Verified {
				status = "✗ verified active"
			}
			fmt.Fprintf(&b, "| `%s` | %d | `%s` | %s |\n",
				fnd.File, fnd.Line, trufflehog.MaskSecret(fnd.Raw), status)
		}
		fmt.Fprintln(&b)

		files := uniqueFileList(matched)
		fmt.Fprintf(&b, "**Affected files**: %d (%s)\n\n", len(files), strings.Join(quoteList(files), ", "))
	}

	fmt.Fprintln(&b, "## Recommended actions")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "1. **즉시** — [%s](%s) 접속해 영향 토큰을 revoke + 새로 발급\n", plat.DashboardURL, plat.DashboardURL)
	fmt.Fprintln(&b, "2. **단기** — 새 토큰을 macOS Keychain에 저장 (`credmux profile create` 또는 `workspace env`)")
	fmt.Fprintln(&b, "3. **운영** — 다음 에이전트 작업은 `credmux workspace open`으로 격리하여 토큰 접근 범위 제한")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "---")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "[← Step 1](./01-exposure.md) · [Next → Step 3](./03-next-actions.md)")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeStep3(path string, plat platform, target string) error {
	var b strings.Builder
	fmt.Fprintln(&b, "# Step 3 — 지속 보호")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "> 사고 대응 후 어떻게 같은 일이 반복되지 않게 할 것인가.")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## 📱 Telegram bot — 비개발자 동료에게도 알림")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```bash")
	fmt.Fprintln(&b, "source .venv-bot/bin/activate")
	fmt.Fprintf(&b, "TELEGRAM_BOT_TOKEN=<your-bot-token> DEMO_HOME=%s python3 bot/bot.py\n", target)
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "스마트폰에서:")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "1. 봇 채팅에 `/start` → 알림 구독")
	fmt.Fprintf(&b, "2. `/testbreach` 입력 → %s 사고 시뮬 알림 + `[✅ 지금 스캔]` 버튼\n", plat.Name)
	fmt.Fprintln(&b, "3. 버튼 탭 → 영향 토큰 한국어 카드 회신 → `[🔑 교체 안내]` 단계별 가이드")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## 🔐 cmux workspace — 다음 에이전트 작업 격리")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```bash")
	fmt.Fprintln(&b, "credmux workspace open <project>")
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "- Core Mode (cmux 미감지): `eval $(credmux workspace env <project>)` 안내")
	fmt.Fprintln(&b, "- Enhanced Mode (cmux 떠 있음):")
	fmt.Fprintln(&b, "  - `workspace.create` 호출")
	fmt.Fprintln(&b, "  - 사이드바: `🔐 <project> | tokens isolated`")
	fmt.Fprintln(&b, "  - 시크릿 watcher 자동 가동 (3초 간격 capture-pane 폴링)")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Operational recommendations")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "- **`.env` 파일은 작업 시작 시 Keychain에서 inject**, 끝나면 unset — 디스크 평문 zero")
	fmt.Fprintln(&b, "- **외부 플랫폼 RSS 구독** — Vercel·Supabase·GitHub 등 사고 발생 시 봇이 자동 알림")
	fmt.Fprintln(&b, "- **에이전트별 workspace 분리** — 한 에이전트가 다른 프로젝트 토큰을 못 보게")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "---")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "[← Step 2](./02-breach-response.md) · [Summary →](./00-summary.md)")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeSummary(path, target string, plat platform, all, matched []trufflehog.Finding, startedAt time.Time) error {
	var b strings.Builder
	fmt.Fprintln(&b, "# credmux audit report")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "**Generated**: %s\n", startedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&b, "**Target**: `%s`\n", target)
	fmt.Fprintf(&b, "**Breach simulation**: %s\n", plat.Name)
	fmt.Fprintln(&b)

	files := uniqueFileList(all)
	fmt.Fprintln(&b, "## Headline numbers")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- **%d** token%s found across **%d** file%s\n",
		len(all), pluralS(len(all)), len(files), pluralS(len(files)))
	fmt.Fprintf(&b, "- **%d** %s token%s would need rotation if %s breached now\n",
		len(matched), plat.Name, pluralS(len(matched)), plat.Name)
	fmt.Fprintln(&b)

	if len(all) > 0 {
		byDet := groupByDetector(all)
		fmt.Fprintln(&b, "### Detector breakdown")
		fmt.Fprintln(&b)
		for _, det := range sortedKeys(byDet) {
			fmt.Fprintf(&b, "- %s — %d\n", det, len(byDet[det]))
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "## Step-by-step report")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "1. [노출 진단](./01-exposure.md) — 환경 전체 토큰 스캔")
	fmt.Fprintf(&b, "2. [%s 사고 대응 시뮬](./02-breach-response.md) — 영향 토큰 + 교체 안내\n", plat.Name)
	fmt.Fprintln(&b, "3. [지속 보호](./03-next-actions.md) — Telegram bot · cmux workspace 격리")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Top 3 immediate actions")
	fmt.Fprintln(&b)
	if len(matched) > 0 {
		fmt.Fprintf(&b, "1. [%s](%s) 접속 → 영향 토큰 %d개 revoke + 새로 발급\n",
			plat.DashboardURL, plat.DashboardURL, len(matched))
		fmt.Fprintln(&b, "2. 새 토큰을 macOS Keychain에 저장 — `.env`에 다시 떨어뜨리지 말 것")
		fmt.Fprintln(&b, "3. 다음 에이전트 작업은 `credmux workspace open`으로 격리하여 사고 재발 방지")
	} else {
		fmt.Fprintln(&b, "1. ✅ 현재 영향 토큰 없음 — 점검 주기 유지")
		fmt.Fprintln(&b, "2. Telegram bot 가동 → 외부 RSS 사고 시 즉시 알림")
		fmt.Fprintln(&b, "3. `credmux workspace open`으로 에이전트별 자격증명 격리")
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "---")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "기존 도구는 git만 본다. credmux는 지금 이 순간을 본다 — 모든 터미널, cmux 안, 비개발자의 Telegram에서.")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "<https://github.com/baekho-lim/credmux>")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func groupByDetector(in []trufflehog.Finding) map[string][]trufflehog.Finding {
	out := map[string][]trufflehog.Finding{}
	for _, f := range in {
		out[f.DetectorName] = append(out[f.DetectorName], f)
	}
	return out
}

func sortedKeys(m map[string][]trufflehog.Finding) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func uniqueFileList(in []trufflehog.Finding) []string {
	seen := map[string]struct{}{}
	for _, f := range in {
		seen[f.File] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func quoteList(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = "`" + s + "`"
	}
	return out
}
