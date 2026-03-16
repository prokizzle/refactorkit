package orchestrator

import (
	"testing"
)

func TestNew(t *testing.T) {
	o, err := New("nextjs")
	if err != nil {
		t.Fatal(err)
	}
	if o.Framework != "nextjs" {
		t.Errorf("expected framework nextjs, got %s", o.Framework)
	}
	if len(o.Skill.Phases) == 0 {
		t.Error("expected phases to be parsed")
	}
	if len(o.References) == 0 {
		t.Error("expected reference documents to be loaded")
	}
	if o.Skill.Name == "" {
		t.Error("expected skill name to be parsed from frontmatter")
	}
}

func TestNew_AllFrameworks(t *testing.T) {
	frameworks := []string{"nextjs", "astro", "ios", "android"}
	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			o, err := New(fw)
			if err != nil {
				t.Fatalf("failed to load %s: %v", fw, err)
			}
			if len(o.Skill.Phases) == 0 {
				t.Errorf("%s: expected phases", fw)
			}
			if len(o.References) == 0 {
				t.Errorf("%s: expected references", fw)
			}
		})
	}
}

func TestNew_InvalidFramework(t *testing.T) {
	_, err := New("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent framework")
	}
}

func TestGetPhase(t *testing.T) {
	o, err := New("nextjs")
	if err != nil {
		t.Fatal(err)
	}
	phase := o.GetPhase("0")
	if phase == nil {
		t.Fatal("expected phase 0 to exist")
	}
	if phase.Title == "" {
		t.Error("expected phase 0 to have a title")
	}
}

func TestGetPhase_NotFound(t *testing.T) {
	o, err := New("nextjs")
	if err != nil {
		t.Fatal(err)
	}
	phase := o.GetPhase("99")
	if phase != nil {
		t.Error("expected nil for nonexistent phase")
	}
}

func TestGetReference(t *testing.T) {
	o, err := New("nextjs")
	if err != nil {
		t.Fatal(err)
	}
	content, ok := o.GetReference("nova-assets.md")
	if !ok {
		t.Error("expected nova-assets.md reference to exist")
	}
	if content == "" {
		t.Error("expected non-empty reference content")
	}
}

func TestListPhases(t *testing.T) {
	o, err := New("nextjs")
	if err != nil {
		t.Fatal(err)
	}
	phases := o.ListPhases()
	if len(phases) < 2 {
		t.Errorf("expected at least 2 phases, got %d", len(phases))
	}
}

func TestSupportedFrameworks(t *testing.T) {
	frameworks := SupportedFrameworks()
	if len(frameworks) < 4 {
		t.Errorf("expected at least 4 frameworks, got %d: %v", len(frameworks), frameworks)
	}
	// Check that expected frameworks are present
	expected := map[string]bool{"nextjs": false, "astro": false, "ios": false, "android": false}
	for _, fw := range frameworks {
		expected[fw] = true
	}
	for fw, found := range expected {
		if !found {
			t.Errorf("expected framework %s to be supported", fw)
		}
	}
}
