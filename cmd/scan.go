package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/baekho-lim/credmux/internal/trufflehog"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a filesystem path for secrets via TruffleHog",
	Long: `scan walks a directory with TruffleHog (700+ patterns) and prints
every finding with the secret masked.

Path resolution: argument › $DEMO_HOME › $HOME › cwd.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func runScan(_ *cobra.Command, args []string) error {
	path := resolveScanPath(args)

	fmt.Printf("🔎 scanning %s …\n\n", path)
	findings, err := trufflehog.Scan(path)
	if err != nil {
		if errors.Is(err, trufflehog.ErrNotInstalled) {
			fmt.Fprintln(os.Stderr, "❌ trufflehog not found in PATH.")
			fmt.Fprintln(os.Stderr, "   Install: brew install trufflehog")
			return err
		}
		return err
	}
	if len(findings) == 0 {
		fmt.Println("✅ No secrets found.")
		return nil
	}
	for _, f := range findings {
		status := "unverified"
		if f.Verified {
			status = "VERIFIED"
		}
		fmt.Printf("  [%s] %-12s %s  %s:%d\n",
			status, f.DetectorName, trufflehog.MaskSecret(f.Raw), f.File, f.Line)
	}
	fmt.Printf("\nTotal: %d finding(s)\n", len(findings))
	return nil
}

func resolveScanPath(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	if d := os.Getenv("DEMO_HOME"); d != "" {
		return d
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}
