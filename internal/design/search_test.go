package design

import (
	"os/exec"
	"testing"
)

func TestSearchRaw(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
	result, err := SearchRaw("modern SaaS dashboard")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
	t.Logf("Search result:\n%s", result[:min(len(result), 500)])
}
