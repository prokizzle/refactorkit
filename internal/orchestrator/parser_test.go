package orchestrator

import (
	"testing"
)

func TestParseSkillMD(t *testing.T) {
	content := `---
name: test-skill
description: Use when testing
---

# Test Skill

## Phase 0 -- Discovery

### 0.1 Detect Framework
Check for marker files.

### 0.2 Run Audit
Generate report.

## Phase 1 -- Infrastructure

### 1.1 Setup
Install dependencies.
`
	skill, err := ParseSkillMD(content)
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "test-skill" {
		t.Errorf("expected name test-skill, got %s", skill.Name)
	}
	if skill.Description != "Use when testing" {
		t.Errorf("expected description 'Use when testing', got %s", skill.Description)
	}
	if len(skill.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(skill.Phases))
	}
	if skill.Phases[0].Number != "0" {
		t.Errorf("expected phase 0, got %s", skill.Phases[0].Number)
	}
	if skill.Phases[0].Title != "Discovery" {
		t.Errorf("expected title Discovery, got %s", skill.Phases[0].Title)
	}
	if len(skill.Phases[0].Sections) != 2 {
		t.Errorf("expected 2 sections in phase 0, got %d", len(skill.Phases[0].Sections))
	}
	if skill.Phases[0].Sections[0].Number != "0.1" {
		t.Errorf("expected section 0.1, got %s", skill.Phases[0].Sections[0].Number)
	}
	if skill.Phases[1].Number != "1" {
		t.Errorf("expected phase 1, got %s", skill.Phases[1].Number)
	}
	if len(skill.Phases[1].Sections) != 1 {
		t.Errorf("expected 1 section in phase 1, got %d", len(skill.Phases[1].Sections))
	}
}

func TestParseSkillMD_EmDash(t *testing.T) {
	content := `---
name: em-dash-skill
description: Use when testing em-dash
---

## Phase 1 — Infrastructure Setup

### 1.1 Setup
Do things.

## Phase 3.5 — Growth & Conversion

### 3.5.1 Marketing
Run campaigns.
`
	skill, err := ParseSkillMD(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(skill.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(skill.Phases))
	}
	if skill.Phases[0].Number != "1" {
		t.Errorf("expected phase 1, got %q", skill.Phases[0].Number)
	}
	if skill.Phases[0].Title != "Infrastructure Setup" {
		t.Errorf("expected title 'Infrastructure Setup', got %q", skill.Phases[0].Title)
	}
	if skill.Phases[1].Number != "3.5" {
		t.Errorf("expected phase 3.5, got %q", skill.Phases[1].Number)
	}
}

func TestParseSkillMD_NoFrontmatter(t *testing.T) {
	content := `## Phase 0 -- Init

### 0.1 Start
Begin.
`
	skill, err := ParseSkillMD(content)
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "" {
		t.Errorf("expected empty name, got %s", skill.Name)
	}
	if len(skill.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(skill.Phases))
	}
}
