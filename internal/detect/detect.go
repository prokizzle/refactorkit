package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Result holds the detected framework information for a project directory.
type Result struct {
	Framework string   `json:"framework"` // "nextjs", "astro", "ios", "android", "unknown"
	Version   string   `json:"version"`
	Features  []string `json:"features"`
}

// Detect inspects the given directory and returns the detected framework.
func Detect(dir string) Result {
	if r := detectNextJS(dir); r.Framework != "unknown" {
		return r
	}
	if r := detectAstro(dir); r.Framework != "unknown" {
		return r
	}
	if r := detectIOS(dir); r.Framework != "unknown" {
		return r
	}
	if r := detectAndroid(dir); r.Framework != "unknown" {
		return r
	}
	return Result{Framework: "unknown"}
}

// packageJSON is a minimal representation of a package.json file.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(dir string) (*packageJSON, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, err
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func hasDep(pkg *packageJSON, name string) (version string, ok bool) {
	if pkg == nil {
		return "", false
	}
	if v, found := pkg.Dependencies[name]; found {
		return v, true
	}
	if v, found := pkg.DevDependencies[name]; found {
		return v, true
	}
	return "", false
}

func anyFileExists(dir string, names ...string) bool {
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func hasFilesWithExt(dir, ext string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found {
			return filepath.SkipAll
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ext) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func hasFileContaining(dir, ext, substring string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found {
			return filepath.SkipAll
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ext) {
			data, readErr := os.ReadFile(path)
			if readErr == nil && strings.Contains(string(data), substring) {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func hasGlobMatch(dir, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	return err == nil && len(matches) > 0
}

// detectNextJS checks for Next.js project markers.
func detectNextJS(dir string) Result {
	configExists := anyFileExists(dir, "next.config.ts", "next.config.js", "next.config.mjs")
	pkg, _ := readPackageJSON(dir)
	version, hasDependency := hasDep(pkg, "next")

	if !configExists && !hasDependency {
		return Result{Framework: "unknown"}
	}

	var features []string

	// app-router: src/app/ directory exists
	if dirExists(filepath.Join(dir, "src", "app")) {
		features = append(features, "app-router")
	}

	// typescript: .ts files exist
	if hasFilesWithExt(dir, ".ts") {
		features = append(features, "typescript")
	}

	// tailwind: tailwindcss in dependencies
	if _, ok := hasDep(pkg, "tailwindcss"); ok {
		features = append(features, "tailwind")
	}

	return Result{
		Framework: "nextjs",
		Version:   version,
		Features:  features,
	}
}

// detectAstro checks for Astro project markers.
func detectAstro(dir string) Result {
	configExists := anyFileExists(dir, "astro.config.ts", "astro.config.js", "astro.config.mjs")
	pkg, _ := readPackageJSON(dir)
	version, hasDependency := hasDep(pkg, "astro")

	if !configExists && !hasDependency {
		return Result{Framework: "unknown"}
	}

	return Result{
		Framework: "astro",
		Version:   version,
	}
}

// detectIOS checks for iOS/Swift project markers.
func detectIOS(dir string) Result {
	hasXcodeProj := hasGlobMatch(dir, "*.xcodeproj")
	hasXcworkspace := hasGlobMatch(dir, "*.xcworkspace")
	hasPackageSwift := anyFileExists(dir, "Package.swift")

	if !hasXcodeProj && !hasXcworkspace && !hasPackageSwift {
		return Result{Framework: "unknown"}
	}

	var features []string

	// swiftui: files importing SwiftUI
	if hasFileContaining(dir, ".swift", "import SwiftUI") {
		features = append(features, "swiftui")
	}

	// tuist: Project.swift exists
	if anyFileExists(dir, "Project.swift") {
		features = append(features, "tuist")
	}

	return Result{
		Framework: "ios",
		Features:  features,
	}
}

// detectAndroid checks for Android project markers.
func detectAndroid(dir string) Result {
	if !anyFileExists(dir, "build.gradle.kts", "build.gradle", "settings.gradle.kts") {
		return Result{Framework: "unknown"}
	}

	var features []string

	// compose: Compose dependencies in gradle files
	if hasFileContaining(dir, ".gradle.kts", "compose") || hasFileContaining(dir, ".gradle", "compose") {
		features = append(features, "compose")
	}

	// kotlin: .kt files exist
	if hasFilesWithExt(dir, ".kt") {
		features = append(features, "kotlin")
	}

	return Result{
		Framework: "android",
		Features:  features,
	}
}
