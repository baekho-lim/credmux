package watcher

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/baekho-lim/credmux/internal/cmux"
)

// Heuristic patterns: shaped like common API tokens. Not exhaustive — this is
// the in-terminal early-warning channel, not the verified-detection layer
// (TruffleHog covers that). False positives are acceptable; a missed leak is
// not.
var patterns = []*regexp.Regexp{
	regexp.MustCompile(`sk-[A-Za-z0-9_\-]{20,}`),       // Anthropic / OpenAI
	regexp.MustCompile(`sk-ant-[A-Za-z0-9_\-]{20,}`),   // Anthropic explicit
	regexp.MustCompile(`ghp_[A-Za-z0-9]{30,}`),         // GitHub PAT
	regexp.MustCompile(`gho_[A-Za-z0-9]{30,}`),         // GitHub OAuth
	regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{10,}`), // Slack
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),             // AWS
	regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),       // GCP
	regexp.MustCompile(`dpl_[A-Za-z0-9]{20,}`),         // Vercel deploy hooks
}

type Config struct {
	SurfaceID   string
	WorkspaceID string
	Project     string
	Interval    time.Duration
}

// Run polls the cmux surface for secret-shaped strings and fires a
// notification + sidebar log on each hit. Stops cleanly when ctx is cancelled.
// All cmux RPC errors are swallowed so a transient socket hiccup does not
// kill the watcher.
func Run(ctx context.Context, cfg Config) {
	if cfg.Interval == 0 {
		cfg.Interval = 3 * time.Second
	}
	t := time.NewTicker(cfg.Interval)
	defer t.Stop()

	var lastHit string
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			text, err := cmux.CapturePane(cfg.SurfaceID)
			if err != nil {
				continue
			}
			hit := match(text)
			if hit == "" || hit == lastHit {
				// Throttle repeated alarms on the same pane snapshot.
				continue
			}
			lastHit = hit
			_ = cmux.CreateNotification(
				"🚨 Secret detected",
				cfg.Project,
				fmt.Sprintf("credmux watcher caught %s in this pane", hit),
			)
			_ = cmux.SidebarLog("error",
				fmt.Sprintf("credmux: secret pattern detected (%s) in workspace %s", hit, cfg.WorkspaceID))
		}
	}
}

func match(s string) string {
	for _, re := range patterns {
		if hit := re.FindString(s); hit != "" {
			return hit
		}
	}
	return ""
}
