package report

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

type Severity string

const (
	Critical Severity = "critical"
	Warning  Severity = "warning"
	Info     Severity = "info"
)

type Issue struct {
	Severity    Severity `json:"severity"`
	Category    string   `json:"category"` // architecture, design, performance, security, scalability
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Description string   `json:"description"`
	Fix         string   `json:"fix"`
}

type OffloadCandidate struct {
	What       string `json:"what"`
	Where      string `json:"where"`
	Why        string `json:"why"`
	CostSaving string `json:"cost_saving"`
}

type ScalingTier struct {
	Users          string `json:"users"`
	CurrentCost    string `json:"current_cost"`
	RefactoredCost string `json:"refactored_cost"`
}

type AuditReport struct {
	Framework         string             `json:"framework"`
	TechDebtScore     int                `json:"tech_debt_score"` // 0-100, 100 = no debt
	InfraTier         string             `json:"infra_tier"`      // lite, standard, full
	Issues            []Issue            `json:"issues"`
	OffloadCandidates []OffloadCandidate `json:"offload_candidates"`
	ScalingTiers      []ScalingTier      `json:"scaling_tiers"`
	GeneratedAt       time.Time          `json:"generated_at"`
}

func (r *AuditReport) CriticalCount() int {
	count := 0
	for _, i := range r.Issues {
		if i.Severity == Critical {
			count++
		}
	}
	return count
}

func (r *AuditReport) WarningCount() int {
	count := 0
	for _, i := range r.Issues {
		if i.Severity == Warning {
			count++
		}
	}
	return count
}

const auditTemplate = `# RefactorKit Audit Report

**Framework:** {{.Framework}}
**Tech Debt Score:** {{.TechDebtScore}}/100
**Infrastructure Tier:** {{.InfraTier}}
**Generated:** {{.GeneratedAt.Format "2006-01-02 15:04"}}

---

## Summary

{{if le .TechDebtScore 30}}Your app has significant technical debt that will block scaling and cause production issues. A full refactoring is strongly recommended.
{{else if le .TechDebtScore 60}}Your app has moderate technical debt. Several issues should be addressed before launch.
{{else if le .TechDebtScore 80}}Your app is in decent shape with some improvements needed.
{{else}}Your app is well-structured. Minor optimizations available.
{{end}}
- **{{.CriticalCount}} critical issues** (must fix before launch)
- **{{.WarningCount}} warnings** (should fix for production quality)
- **{{len .OffloadCandidates}} cloud offloading opportunities**

---

## Issues Found

{{range .Issues}}### [{{.Severity | upper}}] {{.Category}}: {{.Description}}
- **File:** {{.File}}{{if .Line}}:{{.Line}}{{end}}
- **Fix:** {{.Fix}}

{{end}}
{{- if .OffloadCandidates}}
---

## Cloud Offloading Opportunities

These patterns should be moved out of your app:

{{range .OffloadCandidates}}- **{{.What}}** -- {{.Where}}
  - Why: {{.Why}}
  - Estimated savings: {{.CostSaving}}

{{end}}{{end}}
{{- if .ScalingTiers}}
---

## Scaling Cost Projection

| Users | Current Cost | After Refactoring |
|-------|-------------|-------------------|
{{range .ScalingTiers}}| {{.Users}} | {{.CurrentCost}} | {{.RefactoredCost}} |
{{end}}{{end}}
---

## Next Steps

Run ` + "`refactorkit refactor .`" + ` to apply all fixes automatically.
Visit https://refactorkit.dev/pricing for license options.
`

func (r *AuditReport) Render() (string, error) {
	funcMap := template.FuncMap{
		"upper": func(s any) string { return strings.ToUpper(fmt.Sprint(s)) },
		"le":    func(a, b int) bool { return a <= b },
	}
	tmpl, err := template.New("audit").Funcs(funcMap).Parse(auditTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r); err != nil {
		return "", fmt.Errorf("failed to render report: %w", err)
	}
	return buf.String(), nil
}
