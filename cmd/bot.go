package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Telegram bot operations (Component C — SENTINEL)",
}

var botStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Telegram bot (delegates to bot/bot.py)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bot start: run `python3 bot/bot.py` directly for now (Component C — SENTINEL)")
	},
}

func init() {
	botCmd.AddCommand(botStartCmd)
}
