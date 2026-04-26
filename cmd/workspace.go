package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/baekho-lim/credmux/internal/cmux"
	"github.com/baekho-lim/credmux/internal/keychain"
	"github.com/baekho-lim/credmux/internal/watcher"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage isolated agent workspaces",
}

var workspaceOpenCmd = &cobra.Command{
	Use:   "open [project]",
	Short: "Open an isolated workspace with project-scoped credentials",
	Long: `open creates an isolated context for [project].

Core Mode (no cmux): prints a one-liner you can eval to inject the project's
credentials into the current shell.

Enhanced Mode (cmux detected): calls workspace.create, sets the sidebar
status, logs the open event, and (when CMUX_SURFACE_ID is set) starts a
watcher goroutine that polls the surface for secret-shaped strings.`,
	Args: cobra.ExactArgs(1),
	RunE: runWorkspaceOpen,
}

var workspaceEnvCmd = &cobra.Command{
	Use:   "env [project]",
	Short: "Print export-able env vars from Keychain (use with eval $(...))",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkspaceEnv,
}

func runWorkspaceOpen(_ *cobra.Command, args []string) error {
	project := args[0]
	profile, err := keychain.Load(project)
	if err != nil {
		return fmt.Errorf("load profile: %w", err)
	}

	if !cmux.Detect() {
		printCoreFallback(project)
		return nil
	}

	cwd, _ := os.Getwd()
	wsID, err := cmux.CreateWorkspace(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  cmux detected but workspace.create failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "    falling back to Core Mode instructions.")
		printCoreFallback(project)
		return nil
	}

	status := fmt.Sprintf("🔐 %s | tokens isolated", project)
	if err := cmux.SidebarSetStatus(wsID, status, "lock"); err != nil {
		fmt.Fprintf(os.Stderr, "warn: sidebar.set_status failed: %v\n", err)
	}
	if err := cmux.SidebarLog("info",
		fmt.Sprintf("credmux: workspace opened (id=%s project=%s)", wsID, project)); err != nil {
		fmt.Fprintf(os.Stderr, "warn: sidebar.log failed: %v\n", err)
	}

	fmt.Println("✅ Enhanced Mode workspace opened")
	fmt.Printf("   workspace_id: %s\n", wsID)
	fmt.Printf("   project:      %s\n", project)

	injected := keychain.Inject(profile)
	if len(injected) > 0 {
		fmt.Printf("   injected:     %d credential(s) from Keychain\n", len(injected))
	} else {
		fmt.Printf("   injected:     none yet (run breach-drill or `credmux profile create %s` to seed)\n", project)
	}

	surface := os.Getenv("CMUX_SURFACE_ID")
	if surface == "" {
		fmt.Println("   watcher:      skipped — CMUX_SURFACE_ID not set")
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	fmt.Printf("   watcher:      polling surface %s every 3s (Ctrl+C to stop)\n", surface)
	watcher.Run(ctx, watcher.Config{
		SurfaceID:   surface,
		WorkspaceID: wsID,
		Project:     project,
	})
	return nil
}

func runWorkspaceEnv(_ *cobra.Command, args []string) error {
	project := args[0]
	profile, err := keychain.Load(project)
	if err != nil {
		return err
	}
	env := keychain.Inject(profile)
	if len(env) == 0 {
		fmt.Fprintf(os.Stderr, "# credmux: no credentials in Keychain for project %q\n", project)
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("export %s=%q\n", k, env[k])
	}
	return nil
}

func printCoreFallback(project string) {
	fmt.Println("🔐 credmux workspace (Core Mode)")
	fmt.Printf("   project: %s\n", project)
	fmt.Println("   cmux not detected — using shell-injection fallback.")
	fmt.Println()
	fmt.Printf("Inject %s credentials into your current shell:\n", project)
	fmt.Printf("    eval $(credmux workspace env %s)\n", project)
	fmt.Println()
	fmt.Println("To populate the Keychain entries, run:")
	fmt.Println("    credmux breach-drill <platform>")
	fmt.Println("or wire up a profile (planned: credmux profile create).")
}

func init() {
	workspaceCmd.AddCommand(workspaceOpenCmd)
	workspaceCmd.AddCommand(workspaceEnvCmd)
}
