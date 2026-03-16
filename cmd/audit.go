package cmd

import (
	"fmt"

	"github.com/prokizzle/refactorkit/internal/detect"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit [path]",
	Short: "Run a free Phase 0 audit on your codebase",
	Long: `Analyzes your app for tech debt, architecture issues, design gaps,
and scaling cost projections. Always free — no license required.

Examples:
  refactorkit audit .
  refactorkit audit ./my-app --framework nextjs`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		framework, _ := cmd.Flags().GetString("framework")

		// Auto-detect framework if not specified
		if framework == "" {
			result := detect.Detect(path)
			framework = result.Framework
			if framework == "unknown" {
				return fmt.Errorf("could not detect framework in %s — use --framework to specify (nextjs, astro, ios, android)", path)
			}
			fmt.Printf("Detected framework: %s\n", framework)
		}

		// TODO: Phase 0 audit will be implemented in Task 6 (orchestrator) and Task 7 (report generator)
		fmt.Printf("Audit starting for %s project at %s...\n", framework, path)
		fmt.Println("Phase 0 audit is not yet implemented. Coming soon.")
		return nil
	},
}

func init() {
	auditCmd.Flags().StringP("framework", "f", "", "Framework to audit (nextjs, astro, ios, android)")
	rootCmd.AddCommand(auditCmd)
}
