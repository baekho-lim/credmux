package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	auditAuto     bool
	auditPlatform string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit local credentials and walk through breach response",
	Long: `audit는 현재 환경(또는 $DEMO_HOME)에서 노출된 토큰을 찾고, 외부 플랫폼
사고 시나리오를 시뮬레이션하면서 단계별로 다음 액션을 추천하는 인터랙티브
점검이다.

3 단계 — 노출 진단 → 사고 대응 → 지속 보호 — 사이에 [Y/n] 프롬프트가
들어간다 (Enter면 진행).

  --auto      프롬프트 없이 자동 진행 (CI / 에이전트용)
  --platform  사고 대응 시뮬 대상 (vercel|github|stripe|anthropic|supabase)`,
	RunE: runAudit,
}

func init() {
	auditCmd.Flags().BoolVar(&auditAuto, "auto", false,
		"skip Y/n prompts and run all steps non-interactively")
	auditCmd.Flags().StringVar(&auditPlatform, "platform", "vercel",
		"platform to simulate breach response for")
}

type auditStep struct {
	Title string
	Pitch string
	Run   func() error
}

func runAudit(_ *cobra.Command, _ []string) error {
	plat, ok := platforms[strings.ToLower(auditPlatform)]
	if !ok {
		return fmt.Errorf("unknown --platform %q (vercel|github|supabase|stripe|anthropic)", auditPlatform)
	}

	steps := []auditStep{
		{
			Title: "Step 1/3 — 노출 진단",
			Pitch: "이 폴더에 평문으로 떨어진 토큰을 TruffleHog 700+ 패턴으로 찾는다.",
			Run:   auditScanStep,
		},
		{
			Title: fmt.Sprintf("Step 2/3 — %s 사고 대응 시뮬", plat.Name),
			Pitch: fmt.Sprintf("%s가 뚫렸다고 가정. 영향받은 토큰을 즉시 찾고 교체 페이지로 안내.", plat.Name),
			Run:   auditBreachStep,
		},
		{
			Title: "Step 3/3 — 지속 보호",
			Pitch: "비개발자에게는 Telegram 알림으로, 다음 에이전트 작업은 cmux workspace로 격리.",
			Run:   auditNextStep,
		},
	}

	fmt.Println("🔐 credmux audit")
	fmt.Println("   git 이전·이후·런타임의 빈틈을 모두 본다. 에이전트 시대의 자격증명 가디언.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for _, s := range steps {
		fmt.Println("──────────────────────────────────────────────")
		fmt.Println("▶ " + s.Title)
		fmt.Println("  " + s.Pitch)
		fmt.Println()
		if !auditAuto {
			if !askYN(reader, "  계속? [Y/n] ") {
				fmt.Println("  중단됨.")
				return nil
			}
			fmt.Println()
		}
		if err := s.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️  step failed: %v\n", err)
		}
		fmt.Println()
	}

	fmt.Println("──────────────────────────────────────────────")
	fmt.Println("✅ Audit complete.")
	fmt.Println("   기존 도구는 git만 본다. credmux는 지금 이 순간을 본다 —")
	fmt.Println("   모든 터미널에서, cmux 안에서, 비개발자의 Telegram에서.")
	return nil
}

func askYN(r *bufio.Reader, prompt string) bool {
	fmt.Print(prompt)
	line, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes", "예", "ㅇ", "ㅛ":
		return true
	}
	return false
}

func selfExe() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return "./credmux"
}

func auditScanStep() error {
	root := scanRoot()
	fmt.Printf("$ credmux scan %s\n", root)
	c := exec.Command(selfExe(), "scan", root)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func auditBreachStep() error {
	fmt.Printf("$ credmux breach-drill %s\n", auditPlatform)
	c := exec.Command(selfExe(), "breach-drill", auditPlatform)
	c.Env = os.Environ()
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func auditNextStep() error {
	fmt.Println("📱 Telegram (비개발자) — 스마트폰에서:")
	fmt.Println("    /testbreach → [✅ 지금 스캔] → 결과 카드 → [🔑 교체 안내]")
	fmt.Println()
	fmt.Println("🔐 cmux Enhanced (Mac) — 사이드바 격리:")
	fmt.Println("    credmux workspace open <project>")
	fmt.Println("    → 🔐 <project> | tokens isolated  (+ secret watcher 자동 가동)")
	fmt.Println()
	fmt.Println("더 보기: credmux --help")
	return nil
}
