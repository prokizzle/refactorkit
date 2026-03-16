package orchestrator

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed skills
var embeddedSkills embed.FS

// Orchestrator manages the lifecycle of a framework-specific skill orchestration
type Orchestrator struct {
	Skill        *Skill
	Framework    string
	CurrentPhase int
	References   map[string]string // filename -> content
}

// New creates an orchestrator for the given framework by loading
// the embedded SKILL.md and any reference documents.
func New(framework string) (*Orchestrator, error) {
	// Load SKILL.md
	skillPath := filepath.Join("skills", framework, "SKILL.md")
	content, err := fs.ReadFile(embeddedSkills, skillPath)
	if err != nil {
		return nil, fmt.Errorf("no orchestrator skill found for framework %q: %w", framework, err)
	}

	skill, err := ParseSkillMD(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill: %w", err)
	}

	// Load references
	refs := make(map[string]string)
	refDir := filepath.Join("skills", framework, "references")
	entries, _ := fs.ReadDir(embeddedSkills, refDir)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			data, _ := fs.ReadFile(embeddedSkills, filepath.Join(refDir, entry.Name()))
			refs[entry.Name()] = string(data)
		}
	}

	return &Orchestrator{
		Skill:      skill,
		Framework:  framework,
		References: refs,
	}, nil
}

// GetPhase returns a specific phase by its number string (e.g., "0", "0.5", "3.5")
func (o *Orchestrator) GetPhase(number string) *Phase {
	for _, p := range o.Skill.Phases {
		if strings.TrimSpace(p.Number) == number {
			return &p
		}
	}
	return nil
}

// GetReference returns a reference document by filename
func (o *Orchestrator) GetReference(filename string) (string, bool) {
	content, ok := o.References[filename]
	return content, ok
}

// ListPhases returns all phase numbers and titles
func (o *Orchestrator) ListPhases() []string {
	var phases []string
	for _, p := range o.Skill.Phases {
		phases = append(phases, fmt.Sprintf("Phase %s: %s", p.Number, p.Title))
	}
	return phases
}

// SupportedFrameworks returns all frameworks with embedded skills
func SupportedFrameworks() []string {
	var frameworks []string
	entries, _ := fs.ReadDir(embeddedSkills, "skills")
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has a SKILL.md
			if _, err := fs.ReadFile(embeddedSkills, filepath.Join("skills", entry.Name(), "SKILL.md")); err == nil {
				frameworks = append(frameworks, entry.Name())
			}
		}
	}
	return frameworks
}
