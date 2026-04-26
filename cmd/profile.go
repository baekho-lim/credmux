package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage credential profiles (P2 — post-hackathon)",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("profile list: not yet implemented (P2)")
	},
}

var profileCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new credential profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("profile create %s: not yet implemented (P2)\n", args[0])
	},
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCreateCmd)
}
