package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run the bundled hackathon demo (Steps 1-3)",
	Long: `demo runs demo/run_demo.sh from the credmux source tree.

It looks for the script in (in order): the current working directory,
then the directory next to the credmux binary. If you used 'go install'
without cloning the repo, run this from a clone:

  git clone https://github.com/baekho-lim/credmux
  cd credmux
  bash demo/setup.sh && ./credmux demo`,
	RunE: runDemo,
}

func runDemo(_ *cobra.Command, _ []string) error {
	script := findDemoScript()
	if script == "" {
		return fmt.Errorf("demo/run_demo.sh not found in cwd or alongside the binary.\n" +
			"Clone the repo first:\n" +
			"  git clone https://github.com/baekho-lim/credmux && cd credmux && bash demo/setup.sh && ./credmux demo")
	}
	c := exec.Command("bash", script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

func findDemoScript() string {
	candidates := []string{"demo/run_demo.sh"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exe), "demo", "run_demo.sh"))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
