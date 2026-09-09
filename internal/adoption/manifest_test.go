package adoption

import (
	"path/filepath"
	"testing"
)

func TestParseManifestPackageJSON(t *testing.T) {
	dir := t.TempDir()
	content := `{
		"name": "my-web-app",
		"version": "1.2.3",
		"dependencies": {
			"react": "^18.2.0",
			"next": "^14.0.0"
		},
		"devDependencies": {
			"typescript": "^5.0.0",
			"jest": "^29.0.0"
		},
		"scripts": {
			"build": "next build",
			"test": "jest"
		}
	}`
	path := filepath.Join(dir, "package.json")
	writeTestFile(t, dir, "package.json", content)

	pm, err := ParseManifestFile("package.json", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pm.Name != "my-web-app" {
		t.Errorf("expected name 'my-web-app', got %q", pm.Name)
	}
	if len(pm.Dependencies) != 4 {
		t.Errorf("expected 4 dependencies, got %d", len(pm.Dependencies))
	}
	if pm.Scripts["test"] != "jest" {
		t.Errorf("expected script test='jest', got %q", pm.Scripts["test"])
	}
}

func TestParseManifestGoMod(t *testing.T) {
	dir := t.TempDir()
	content := `module github.com/example/service

go 1.22

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
)

require github.com/lib/pq v1.10.9
`
	path := filepath.Join(dir, "go.mod")
	writeTestFile(t, dir, "go.mod", content)

	pm, err := ParseManifestFile("go.mod", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pm.Name != "github.com/example/service" {
		t.Errorf("expected module name, got %q", pm.Name)
	}
	if len(pm.Dependencies) != 3 {
		t.Errorf("expected 3 dependencies, got %d", len(pm.Dependencies))
	}
}

func TestParseManifestPyprojectTOML(t *testing.T) {
	dir := t.TempDir()
	content := `[project]
name = "api-service"
version = "0.1.0"
dependencies = [
    "fastapi>=0.100.0",
    "uvicorn[standard]",
    "sqlalchemy",
]
`
	path := filepath.Join(dir, "pyproject.toml")
	writeTestFile(t, dir, "pyproject.toml", content)

	pm, err := ParseManifestFile("pyproject.toml", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pm.Name != "api-service" {
		t.Errorf("expected name 'api-service', got %q", pm.Name)
	}
	if len(pm.Dependencies) != 3 {
		t.Errorf("expected 3 dependencies, got %d (%v)", len(pm.Dependencies), pm.Dependencies)
	}
}

func TestParseManifestCargoTOML(t *testing.T) {
	dir := t.TempDir()
	content := `[package]
name = "cli-tool"
version = "0.1.0"

[dependencies]
clap = { version = "4.0", features = ["derive"] }
tokio = "1.0"
`
	path := filepath.Join(dir, "Cargo.toml")
	writeTestFile(t, dir, "Cargo.toml", content)

	pm, err := ParseManifestFile("Cargo.toml", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pm.Name != "cli-tool" {
		t.Errorf("expected name 'cli-tool', got %q", pm.Name)
	}
	if len(pm.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d (%v)", len(pm.Dependencies), pm.Dependencies)
	}
}
