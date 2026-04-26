package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/baekho/credmux/internal/cmux"
	"github.com/baekho/credmux/internal/trufflehog"
	"github.com/spf13/cobra"
)

type platform struct {
	Name         string
	DetectorName string
	DashboardURL string
}

var platforms = map[string]platform{
	"vercel":    {"Vercel", "Vercel", "https://vercel.com/account/tokens"},
	"github":    {"GitHub", "GitHub", "https://github.com/settings/tokens"},
	"supabase":  {"Supabase", "Supabase", "https://app.supabase.com/account/tokens"},
	"stripe":    {"Stripe", "Stripe", "https://dashboard.stripe.com/apikeys"},
	"anthropic": {"Anthropic", "Anthropic", "https://console.anthropic.com/account/keys"},
}

var breachJSON bool

var breachCmd = &cobra.Command{
	Use:   "breach-drill [platform]",
	Short: "Scan for tokens of a breached platform and guide rotation",
	Long: `breach-drill scans the local filesystem for tokens belonging to a
platform that has just disclosed a security incident, then points you at the
correct rotation URL.

Scan root resolution: $DEMO_HOME › $HOME › cwd.
Supported platforms: vercel, github, supabase, stripe, anthropic.

--json emits a stable machine-readable payload (consumed by the Telegram bot).
In JSON mode every error path also produces valid JSON with the "error" field
populated, so callers can rely on parsing stdout regardless of exit code.`,
	Args: cobra.ExactArgs(1),
	RunE: runBreachDrill,
}

func init() {
	breachCmd.Flags().BoolVar(&breachJSON, "json", false, "emit machine-readable JSON to stdout")
}

type breachFinding struct {
	Detector  string `json:"detector"`
	File      string `json:"file"`
	Line      int64  `json:"line"`
	RawMasked string `json:"raw_masked"`
	Verified  bool   `json:"verified"`
}

type breachReport struct {
	Platform     string          `json:"platform"`
	PlatformName string          `json:"platform_name,omitempty"`
	DashboardURL string          `json:"dashboard_url,omitempty"`
	ScanRoot     string          `json:"scan_root,omitempty"`
	Findings     []breachFinding `json:"findings"`
	ImpactFiles  int             `json:"impact_files"`
	Error        string          `json:"error,omitempty"`
}

func emitJSON(r breachReport) error {
	if r.Findings == nil {
		r.Findings = []breachFinding{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func runBreachDrill(_ *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])
	p, ok := platforms[key]
	if !ok {
		if breachJSON {
			return emitJSON(breachReport{
				Platform: key,
				Error:    fmt.Sprintf("unknown platform %q", key),
			})
		}
		return fmt.Errorf("unknown platform %q (supported: vercel, github, supabase, stripe, anthropic)", key)
	}

	root := scanRoot()

	if !breachJSON {
		fmt.Printf("🚨 %s Security Drill\n", p.Name)
		fmt.Printf("   scanning %s (TruffleHog 700+ patterns)\n\n", root)
	}

	findings, err := trufflehog.Scan(root)
	if err != nil {
		if breachJSON {
			return emitJSON(breachReport{
				Platform:     key,
				PlatformName: p.Name,
				DashboardURL: p.DashboardURL,
				ScanRoot:     root,
				Error:        err.Error(),
			})
		}
		if errors.Is(err, trufflehog.ErrNotInstalled) {
			fmt.Fprintln(os.Stderr, "❌ trufflehog not found in PATH.")
			fmt.Fprintln(os.Stderr, "   Install: brew install trufflehog")
			return err
		}
		fmt.Fprintf(os.Stderr, "⚠️  scan finished with error: %v\n", err)
	}

	matched := filterByDetector(findings, p.DetectorName)

	if breachJSON {
		report := breachReport{
			Platform:     key,
			PlatformName: p.Name,
			DashboardURL: p.DashboardURL,
			ScanRoot:     root,
			Findings:     make([]breachFinding, 0, len(matched)),
			ImpactFiles:  uniqueFiles(matched),
		}
		for _, f := range matched {
			report.Findings = append(report.Findings, breachFinding{
				Detector:  f.DetectorName,
				File:      f.File,
				Line:      f.Line,
				RawMasked: trufflehog.MaskSecret(f.Raw),
				Verified:  f.Verified,
			})
		}
		return emitJSON(report)
	}

	if len(matched) == 0 {
		fmt.Println("✅ No matching tokens found. You're safe.")
		fmt.Printf("→ %s\n", p.DashboardURL)
		return nil
	}

	fmt.Printf("Found %d %s token(s):\n", len(matched), p.Name)
	for _, f := range matched {
		status := "⚠️  unverified"
		if f.Verified {
			status = "✗  verified active"
		}
		fmt.Printf("  %s:%d\n    %s  %s  %s\n",
			f.File, f.Line, f.DetectorName, trufflehog.MaskSecret(f.Raw), status)
	}
	fmt.Printf("\nImpact: %d file(s)\n", uniqueFiles(matched))

	fmt.Printf("→ Rotate at: %s\n", p.DashboardURL)
	if cmux.Detect() {
		if err := cmux.OpenBrowser(p.DashboardURL); err == nil {
			fmt.Println("  (opened in cmux browser)")
		} else {
			fmt.Fprintf(os.Stderr, "  (cmux browser.open failed: %v)\n", err)
		}
	}
	return nil
}

func scanRoot() string {
	if d := os.Getenv("DEMO_HOME"); d != "" {
		return d
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}

func filterByDetector(in []trufflehog.Finding, detector string) []trufflehog.Finding {
	var out []trufflehog.Finding
	for _, f := range in {
		if strings.EqualFold(f.DetectorName, detector) {
			out = append(out, f)
		}
	}
	return out
}

func uniqueFiles(in []trufflehog.Finding) int {
	seen := map[string]struct{}{}
	for _, f := range in {
		seen[f.File] = struct{}{}
	}
	return len(seen)
}
