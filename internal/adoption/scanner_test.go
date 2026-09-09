package adoption

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestScanMinimalRepo(t *testing.T) {
	root := t.TempDir()

	// Create a minimal repo structure.
	writeTestFile(t, root, "go.mod", "module example.com/demo\n\ngo 1.22\n")
	writeTestFile(t, root, "main.go", "package main\n\nfunc main() {}\n")
	writeTestFile(t, root, "README.md", "# Demo\n\nA test project.\n")
	writeTestFile(t, root, "Makefile", "build:\n\tgo build .\n")
	mkdirTest(t, root, "docs")
	writeTestFile(t, root, "docs/design.md", "# Design\n\nArchitecture.\n")
	mkdirTest(t, root, ".github/workflows")
	writeTestFile(t, root, ".github/workflows/ci.yml", "name: CI\non: push\n")
	mkdirTest(t, root, "tests")
	writeTestFile(t, root, "tests/main_test.go", "package tests\n")
	writeTestFile(t, root, "AGENTS.md", "# Agent Rules\n")

	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	if result.Version != 1 {
		t.Fatalf("expected version 1, got %d", result.Version)
	}
	if len(result.Facts) == 0 {
		t.Fatal("expected facts, got none")
	}

	// Verify fact types are present.
	kinds := map[FactKind]bool{}
	for _, f := range result.Facts {
		kinds[f.Kind] = true
		if err := f.Validate(); err != nil {
			t.Errorf("invalid fact %s: %v", f.ID, err)
		}
	}
	for _, expected := range []FactKind{FactManifest, FactFile, FactDoc, FactCI, FactTest, FactAgentRules} {
		if !kinds[expected] {
			t.Errorf("expected fact kind %q not found in scan results", expected)
		}
	}

	// Verify JSON serialisation round-trips.
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	var decoded ScanResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if len(decoded.Facts) != len(result.Facts) {
		t.Fatalf("fact count mismatch after round-trip: %d vs %d", len(decoded.Facts), len(result.Facts))
	}
}

func TestScanBudgetExhaustion(t *testing.T) {
	root := t.TempDir()
	mkdirTest(t, root, "src")
	for i := 0; i < 20; i++ {
		name := "file" + string(rune('a'+i)) + ".go"
		writeTestFile(t, root, filepath.Join("src", name), "package src\n")
	}

	scanner := NewScanner(root, ScanOptions{Budget: ScanBudget{MaxFiles: 5, MaxBytes: 50 * 1024 * 1024}})
	result := scanner.Scan()

	if !result.Budget.Exhausted {
		t.Fatal("expected budget to be exhausted")
	}
	if result.Budget.FilesScanned > 5 {
		t.Fatalf("scanned %d files, expected at most 5", result.Budget.FilesScanned)
	}
}

func TestScanBytesBudgetExhaustion(t *testing.T) {
	root := t.TempDir()
	// Create a file that exceeds a very small byte budget.
	writeTestFile(t, root, "big.go", string(make([]byte, 1024)))

	scanner := NewScanner(root, ScanOptions{Budget: ScanBudget{MaxFiles: 1000, MaxBytes: 100}})
	result := scanner.Scan()

	if !result.Budget.Exhausted {
		t.Fatal("expected byte budget to be exhausted")
	}
}

func TestScanExclusions(t *testing.T) {
	root := t.TempDir()
	mkdirTest(t, root, "node_modules/pkg")
	writeTestFile(t, root, "node_modules/pkg/index.js", "module.exports = {}\n")
	mkdirTest(t, root, "vendor/lib")
	writeTestFile(t, root, "vendor/lib/dep.go", "package lib\n")
	writeTestFile(t, root, "main.go", "package main\n")

	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	for _, f := range result.Facts {
		if f.Source == filepath.Join("node_modules", "pkg", "index.js") ||
			f.Source == filepath.Join("vendor", "lib", "dep.go") {
			t.Errorf("excluded path %q should not appear in facts", f.Source)
		}
	}

	// main.go should still be present.
	found := false
	for _, f := range result.Facts {
		if f.Source == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("main.go should appear in scan results")
	}
}

func TestScanCustomExclusions(t *testing.T) {
	root := t.TempDir()
	mkdirTest(t, root, "generated")
	writeTestFile(t, root, "generated/auto.go", "package generated\n")
	writeTestFile(t, root, "main.go", "package main\n")

	scanner := NewScanner(root, ScanOptions{
		Budget:          DefaultBudget(),
		ExcludePatterns: []string{"generated"},
	})
	result := scanner.Scan()

	for _, f := range result.Facts {
		if f.Source == filepath.Join("generated", "auto.go") {
			t.Errorf("custom-excluded path should not appear in facts")
		}
	}
}

func TestScanEmptyRepo(t *testing.T) {
	root := t.TempDir()
	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	if result.Version != 1 {
		t.Fatalf("expected version 1, got %d", result.Version)
	}
	if len(result.Facts) != 0 {
		t.Fatalf("expected no facts, got %d", len(result.Facts))
	}
	if result.Budget.Exhausted {
		t.Fatal("budget should not be exhausted for empty repo")
	}
}

func TestScanHashDeterminism(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "main.go", "package main\n")

	s1 := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	r1 := s1.Scan()
	s2 := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	r2 := s2.Scan()

	if len(r1.Facts) != len(r2.Facts) {
		t.Fatal("scan results differ in fact count")
	}
	for i := range r1.Facts {
		if r1.Facts[i].SourceHash != r2.Facts[i].SourceHash {
			t.Errorf("hash mismatch for %s", r1.Facts[i].Source)
		}
	}
}

func TestScanManifestClassification(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "package.json", `{"name":"demo"}`)
	writeTestFile(t, root, "go.mod", "module demo\n\ngo 1.22\n")
	writeTestFile(t, root, "pyproject.toml", "[project]\nname=\"demo\"\n")
	writeTestFile(t, root, "Cargo.toml", "[package]\nname=\"demo\"\n")

	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	manifestCount := 0
	for _, f := range result.Facts {
		if f.Kind == FactManifest {
			manifestCount++
		}
	}
	if manifestCount < 4 {
		t.Fatalf("expected at least 4 manifest facts, got %d", manifestCount)
	}
}

func TestScanSchemaClassification(t *testing.T) {
	root := t.TempDir()
	mkdirTest(t, root, "schemas")
	writeTestFile(t, root, "schemas/user.schema.json", `{"type":"object"}`)
	writeTestFile(t, root, "api.graphql", "type Query { hello: String }")

	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	schemaCount := 0
	for _, f := range result.Facts {
		if f.Kind == FactSchema {
			schemaCount++
		}
	}
	if schemaCount < 2 {
		t.Fatalf("expected at least 2 schema facts, got %d", schemaCount)
	}
}

func TestScanConfigClassification(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".gitignore", "*.o\n")
	writeTestFile(t, root, "Makefile", "all: build\n")
	writeTestFile(t, root, "Dockerfile", "FROM golang:1.22\n")

	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	configCount := 0
	for _, f := range result.Facts {
		if f.Kind == FactConfig {
			configCount++
		}
	}
	if configCount < 3 {
		t.Fatalf("expected at least 3 config facts, got %d", configCount)
	}
}

func TestScanFactIDsAreUnique(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "a.go", "package a\n")
	writeTestFile(t, root, "b.go", "package b\n")
	writeTestFile(t, root, "c.go", "package c\n")

	scanner := NewScanner(root, ScanOptions{Budget: DefaultBudget()})
	result := scanner.Scan()

	ids := map[string]bool{}
	for _, f := range result.Facts {
		if ids[f.ID] {
			t.Fatalf("duplicate fact ID: %s", f.ID)
		}
		ids[f.ID] = true
	}
}

// --- helpers ---

func writeTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mkdirTest(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, rel), 0755); err != nil {
		t.Fatal(err)
	}
}
