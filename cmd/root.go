package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "credmux",
	Short: "Security guardian for vibe-coders in the AI agent era",
	Long: `credmux — works in any terminal, gains superpowers in cmux.

Core Mode runs everywhere: scan, breach-drill, profile, telegram bot.
Enhanced Mode auto-activates when $CMUX_SOCKET_PATH is detected.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(breachCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(botCmd)
	rootCmd.AddCommand(auditCmd)
}
