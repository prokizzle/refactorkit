package cmd

import (
	"fmt"
	"os"

	"github.com/prokizzle/refactorkit/internal/auth"
	"github.com/prokizzle/refactorkit/internal/detect"
	"github.com/spf13/cobra"
)

var refactorCmd = &cobra.Command{
	Use:   "refactor [path]",
	Short: "Transform a vibe-coded app into production-grade software",
	Long: `Runs the full phased orchestrator: infrastructure, architecture, design system,
monetization, observability, and launch readiness.

Requires a valid license. Run 'refactorkit audit' first for a free assessment.

Examples:
  refactorkit refactor .
  refactorkit refactor . --framework nextjs
  refactorkit refactor . --provider anthropic --model claude-sonnet-4-6`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		framework, _ := cmd.Flags().GetString("framework")
		licenseKey, _ := cmd.Flags().GetString("license")

		// Check for license key in flag, env, or config
		if licenseKey == "" {
			licenseKey = os.Getenv("REFACTORKIT_LICENSE")
		}
		if licenseKey == "" {
			fmt.Println("License required for full refactoring.")
			fmt.Println("Run 'refactorkit audit .' for a free assessment.")
			fmt.Println("Get a license at https://refactorkit.dev/pricing")
			return nil
		}

		// Validate license
		status, err := auth.ValidateLicense(licenseKey)
		if err != nil || !status.CanRefactor() {
			fmt.Println("Invalid or expired license. Visit https://refactorkit.dev/pricing")
			return nil
		}

		fmt.Printf("License valid: %s tier, %d projects remaining\n", status.Tier, status.ProjectsRemaining)

		// Auto-detect framework
		if framework == "" {
			result := detect.Detect(path)
			framework = result.Framework
			if framework == "unknown" {
				return fmt.Errorf("could not detect framework — use --framework to specify")
			}
			fmt.Printf("Detected framework: %s\n", framework)
		}

		// TODO: Full orchestrator will be implemented in Task 6
		fmt.Printf("Refactoring %s project at %s...\n", framework, path)
		fmt.Println("Full orchestrator is not yet implemented. Coming soon.")
		return nil
	},
}

func init() {
	refactorCmd.Flags().StringP("framework", "f", "", "Framework (nextjs, astro, ios, android)")
	refactorCmd.Flags().StringP("provider", "p", "", "AI provider (anthropic, openai, google, groq)")
	refactorCmd.Flags().StringP("model", "m", "", "AI model to use")
	refactorCmd.Flags().StringP("license", "l", "", "License key (or set REFACTORKIT_LICENSE env var)")
	rootCmd.AddCommand(refactorCmd)
}
