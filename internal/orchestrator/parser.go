package orchestrator

import (
	"bufio"
	"strings"
)

// Skill represents a parsed SKILL.md file
type Skill struct {
	Name        string
	Description string
	RawContent  string
	Phases      []Phase
}

// Phase represents a major phase in the orchestration flow
type Phase struct {
	Number   string // "0", "0.5", "1", "2", etc.
	Title    string // "Discovery & Research"
	Sections []Section
	RawContent string
}

// Section represents a subsection within a phase
type Section struct {
	Number  string // "0.1", "1.2", etc.
	Title   string
	Content string
}

// ParseSkillMD parses a SKILL.md file into structured phases.
// It handles YAML frontmatter, phase headers (## Phase N), and
// section headers (### N.N Title). Supports both em-dash and double-dash separators.
func ParseSkillMD(content string) (*Skill, error) {
	skill := &Skill{RawContent: content}

	scanner := bufio.NewScanner(strings.NewReader(content))
	inFrontmatter := false
	frontmatterDone := false // once closed, never re-enter
	lineNum := 0
	var currentPhase *Phase
	var currentSection *Section
	var contentBuilder strings.Builder
	var phaseContentBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Parse YAML frontmatter (only at the start of the file)
		if line == "---" && !frontmatterDone {
			if !inFrontmatter {
				// Only treat as frontmatter opener if it's the first line
				if lineNum == 1 {
					inFrontmatter = true
				}
				continue
			}
			inFrontmatter = false
			frontmatterDone = true
			continue
		}
		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				skill.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "description:") {
				skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
			continue
		}

		// Detect phase headers: "## Phase N" or "## Phase N.N"
		if strings.HasPrefix(line, "## Phase ") || strings.HasPrefix(line, "## phase ") {
			// Save previous phase
			if currentPhase != nil {
				if currentSection != nil {
					currentSection.Content = contentBuilder.String()
					currentPhase.Sections = append(currentPhase.Sections, *currentSection)
					currentSection = nil
				}
				currentPhase.RawContent = phaseContentBuilder.String()
				skill.Phases = append(skill.Phases, *currentPhase)
			}
			contentBuilder.Reset()
			phaseContentBuilder.Reset()

			// Normalize separators: em-dash, en-dash, double-dash
			normalized := line
			normalized = strings.ReplaceAll(normalized, "\u2014", "|") // em-dash
			normalized = strings.ReplaceAll(normalized, "\u2013", "|") // en-dash
			normalized = strings.ReplaceAll(normalized, "--", "|")     // double-dash

			parts := strings.SplitN(normalized, "|", 2)
			// Extract phase number from "## Phase 0.5"
			phasePrefix := strings.TrimPrefix(parts[0], "## Phase ")
			phasePrefix = strings.TrimPrefix(phasePrefix, "## phase ")
			number := strings.TrimSpace(phasePrefix)

			title := ""
			if len(parts) > 1 {
				title = strings.TrimSpace(parts[1])
			}
			currentPhase = &Phase{Number: number, Title: title}
			continue
		}

		// Detect section headers: "### N.N Title"
		if strings.HasPrefix(line, "### ") && currentPhase != nil {
			if currentSection != nil {
				currentSection.Content = contentBuilder.String()
				currentPhase.Sections = append(currentPhase.Sections, *currentSection)
				contentBuilder.Reset()
			}
			header := strings.TrimPrefix(line, "### ")
			spaceIdx := strings.Index(header, " ")
			number := header
			title := ""
			if spaceIdx > 0 {
				number = header[:spaceIdx]
				title = header[spaceIdx+1:]
			}
			currentSection = &Section{Number: number, Title: title}
			phaseContentBuilder.WriteString(line + "\n")
			continue
		}

		if currentSection != nil {
			contentBuilder.WriteString(line + "\n")
		}
		if currentPhase != nil {
			phaseContentBuilder.WriteString(line + "\n")
		}
	}

	// Don't forget the last phase/section
	if currentSection != nil && currentPhase != nil {
		currentSection.Content = contentBuilder.String()
		currentPhase.Sections = append(currentPhase.Sections, *currentSection)
	}
	if currentPhase != nil {
		currentPhase.RawContent = phaseContentBuilder.String()
		skill.Phases = append(skill.Phases, *currentPhase)
	}

	return skill, nil
}
