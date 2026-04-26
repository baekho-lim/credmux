package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var demoAuto bool

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run the 1-minute interactive demo",
	Long: `demo는 1분 시연용 인터랙티브 walkthrough다.
3개 단계 — 평문 노출 → breach-drill → 다음 액션 — 사이에 [Y/n] 프롬프트가
들어간다. --auto 플래그로 프롬프트 없이 자동 진행.`,
	RunE: runDemo,
}

func init() {
	demoCmd.Flags().BoolVar(&demoAuto, "auto", false,
		"skip Y/n prompts and run all steps non-interactively")
}

type demoStep struct {
	Title string
	Pitch string
	Run   func() error
}

func runDemo(_ *cobra.Command, _ []string) error {
	steps := []demoStep{
		{
			Title: "Step 1/3 — 평문 토큰 노출",
			Pitch: "AI 에이전트가 .env를 그대로 읽는다. git 이전의 빈틈이다.",
			Run:   demoStep1,
		},
		{
			Title: "Step 2/3 — breach-drill (Vercel 사고 시뮬)",
			Pitch: "사고 발생 → 영향받은 토큰 즉시 탐지 + 교체 페이지 안내. ~10초.",
			Run:   demoStep2,
		},
		{
			Title: "Step 3/3 — 다음 액션 (격리 / 알림)",
			Pitch: "비개발자 동료는 Telegram으로, 다음 에이전트 작업은 cmux workspace로.",
			Run:   demoStep3,
		},
	}

	fmt.Println("🔐 credmux — 1-minute demo")
	fmt.Println("   git 이전·이후·런타임의 빈틈을 모두 본다. 에이전트 시대의 보안 가디언.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for _, s := range steps {
		fmt.Println("──────────────────────────────────────────────")
		fmt.Println("▶ " + s.Title)
		fmt.Println("  " + s.Pitch)
		fmt.Println()
		if !demoAuto {
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
	fmt.Println("✅ Done.")
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

func demoStep1() error {
	fmt.Println("$ cat demo/shop/.env | grep TOKEN")
	c := exec.Command("bash", "-c", "cat demo/shop/.env | grep TOKEN")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func demoStep2() error {
	fmt.Println("$ DEMO_HOME=./demo credmux breach-drill vercel")
	exe, err := os.Executable()
	if err != nil {
		exe = "./credmux"
	}
	c := exec.Command(exe, "breach-drill", "vercel")
	c.Env = append(os.Environ(), "DEMO_HOME=./demo")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func demoStep3() error {
	fmt.Println("📱 Telegram (비개발자) — 스마트폰에서:")
	fmt.Println("    /testbreach → [✅ 지금 스캔] → 결과 카드 → [🔑 교체 안내]")
	fmt.Println()
	fmt.Println("🔐 cmux Enhanced (Mac) — 사이드바 격리:")
	fmt.Println("    ./credmux workspace open demo")
	fmt.Println("    → 🔐 demo | tokens isolated  (+ secret watcher 자동 가동)")
	return nil
}
