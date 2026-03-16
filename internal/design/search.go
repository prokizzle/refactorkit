package design

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed data/*.csv
var designData embed.FS

//go:embed scripts/design-search.py
var searchScript []byte

type SearchResult struct {
	Domain  string            `json:"domain"`
	Query   string            `json:"query"`
	Results []map[string]string `json:"results"`
}

// Search runs the BM25 design search with the given query
func Search(query string) ([]SearchResult, error) {
	tmpDir, err := os.MkdirTemp("", "refactorkit-design-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}
	entries, err := designData.ReadDir("data")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded data dir: %w", err)
	}
	for _, entry := range entries {
		content, err := designData.ReadFile("data/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded file %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dataDir, entry.Name()), content, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", entry.Name(), err)
		}
	}

	scriptDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create scripts dir: %w", err)
	}
	scriptPath := filepath.Join(scriptDir, "design-search.py")
	if err := os.WriteFile(scriptPath, searchScript, 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	cmd := exec.Command("python3", scriptPath, query)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("design search failed: %w\nOutput: %s", err, string(output))
	}

	// The Python script outputs markdown by default, not structured JSON.
	// For structured results, use SearchRaw and parse as needed.
	_ = output
	return nil, nil
}

// SearchRaw runs the search and returns raw markdown output
func SearchRaw(query string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "refactorkit-design-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data dir: %w", err)
	}
	entries, err := designData.ReadDir("data")
	if err != nil {
		return "", fmt.Errorf("failed to read embedded data dir: %w", err)
	}
	for _, entry := range entries {
		content, err := designData.ReadFile("data/" + entry.Name())
		if err != nil {
			return "", fmt.Errorf("failed to read embedded file %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dataDir, entry.Name()), content, 0644); err != nil {
			return "", fmt.Errorf("failed to write file %s: %w", entry.Name(), err)
		}
	}

	scriptDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create scripts dir: %w", err)
	}
	scriptPath := filepath.Join(scriptDir, "design-search.py")
	if err := os.WriteFile(scriptPath, searchScript, 0644); err != nil {
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	cmd := exec.Command("python3", scriptPath, query)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("design search failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
