package report

import (
	"strings"
	"testing"
	"time"
)

func TestAuditReport_Render(t *testing.T) {
	report := &AuditReport{
		Framework:     "nextjs",
		TechDebtScore: 35,
		InfraTier:     "standard",
		Issues: []Issue{
			{Severity: Critical, Category: "architecture", File: "src/app/api/upload/route.ts", Line: 15, Description: "Image processing blocks event loop", Fix: "Move to AWS Lambda with Sharp layer"},
			{Severity: Warning, Category: "design", File: "src/app/page.tsx", Line: 1, Description: "Default Tailwind colors used", Fix: "Apply custom design tokens from design system"},
			{Severity: Warning, Category: "performance", File: "src/app/layout.tsx", Line: 30, Description: "No error boundary", Fix: "Add global-error.tsx for App Router"},
		},
		OffloadCandidates: []OffloadCandidate{
			{What: "Image processing with sharp", Where: "AWS Lambda + S3 trigger", Why: "Blocks event loop, times out at 10s on Vercel", CostSaving: "$0.20/1K vs crashes at 50 concurrent"},
		},
		ScalingTiers: []ScalingTier{
			{Users: "1K", CurrentCost: "$0", RefactoredCost: "$0"},
			{Users: "10K", CurrentCost: "$75", RefactoredCost: "$50"},
			{Users: "100K", CurrentCost: "Crashes", RefactoredCost: "$400"},
		},
		GeneratedAt: time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC),
	}

	output, err := report.Render()
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	// Verify key content
	checks := []string{
		"Tech Debt Score:** 35/100",
		"nextjs",
		"[CRITICAL]",
		"Image processing blocks event loop",
		"Cloud Offloading",
		"AWS Lambda",
		"Scaling Cost Projection",
		"refactorkit refactor",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing: %q", check)
		}
	}
}

func TestAuditReport_Counts(t *testing.T) {
	report := &AuditReport{
		Issues: []Issue{
			{Severity: Critical}, {Severity: Critical},
			{Severity: Warning}, {Severity: Info},
		},
	}
	if report.CriticalCount() != 2 {
		t.Errorf("expected 2 critical")
	}
	if report.WarningCount() != 1 {
		t.Errorf("expected 1 warning")
	}
}
