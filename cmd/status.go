package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the progress of a running or paused refactoring session",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Read from session database
		fmt.Println("No active refactoring session found.")
		fmt.Println("Start one with: refactorkit refactor .")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
