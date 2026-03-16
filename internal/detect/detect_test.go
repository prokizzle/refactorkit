package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectNextJS(t *testing.T) {
	dir := t.TempDir()

	// Create next.config.ts
	if err := os.WriteFile(filepath.Join(dir, "next.config.ts"), []byte("export default {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create package.json with next dependency
	pkgJSON := `{"dependencies":{"next":"14.2.0","react":"18.2.0"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	r := Detect(dir)
	if r.Framework != "nextjs" {
		t.Errorf("expected framework 'nextjs', got %q", r.Framework)
	}
	if r.Version != "14.2.0" {
		t.Errorf("expected version '14.2.0', got %q", r.Version)
	}
}

func TestDetectAstro(t *testing.T) {
	dir := t.TempDir()

	// Create astro.config.mjs
	if err := os.WriteFile(filepath.Join(dir, "astro.config.mjs"), []byte("export default {}"), 0644); err != nil {
		t.Fatal(err)
	}

	r := Detect(dir)
	if r.Framework != "astro" {
		t.Errorf("expected framework 'astro', got %q", r.Framework)
	}
}

func TestDetectIOS(t *testing.T) {
	dir := t.TempDir()

	// Create a fake .xcodeproj directory
	if err := os.Mkdir(filepath.Join(dir, "MyApp.xcodeproj"), 0755); err != nil {
		t.Fatal(err)
	}

	r := Detect(dir)
	if r.Framework != "ios" {
		t.Errorf("expected framework 'ios', got %q", r.Framework)
	}
}

func TestDetectAndroid(t *testing.T) {
	dir := t.TempDir()

	// Create build.gradle.kts
	if err := os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte("plugins { id(\"com.android.application\") }"), 0644); err != nil {
		t.Fatal(err)
	}

	r := Detect(dir)
	if r.Framework != "android" {
		t.Errorf("expected framework 'android', got %q", r.Framework)
	}
}

func TestDetectUnknown(t *testing.T) {
	dir := t.TempDir()

	r := Detect(dir)
	if r.Framework != "unknown" {
		t.Errorf("expected framework 'unknown', got %q", r.Framework)
	}
}

func TestDetectFeatures(t *testing.T) {
	dir := t.TempDir()

	// Create next.config.ts
	if err := os.WriteFile(filepath.Join(dir, "next.config.ts"), []byte("export default {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create package.json with next and tailwindcss
	pkgJSON := `{"dependencies":{"next":"15.0.0","tailwindcss":"3.4.0"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create src/app/ directory for app-router detection
	if err := os.MkdirAll(filepath.Join(dir, "src", "app"), 0755); err != nil {
		t.Fatal(err)
	}

	r := Detect(dir)
	if r.Framework != "nextjs" {
		t.Fatalf("expected framework 'nextjs', got %q", r.Framework)
	}

	features := make(map[string]bool)
	for _, f := range r.Features {
		features[f] = true
	}

	if !features["app-router"] {
		t.Error("expected 'app-router' feature to be detected")
	}
	if !features["tailwind"] {
		t.Error("expected 'tailwind' feature to be detected")
	}
}
