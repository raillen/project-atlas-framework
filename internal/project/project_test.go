package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "prumo.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create prumo.json: %v", err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	found, err := FindRoot(nested)
	if err != nil {
		t.Fatalf("expected root to be found, got error: %v", err)
	}
	if found != root {
		t.Fatalf("expected %s, got %s", root, found)
	}
}

func TestFindRootNotFound(t *testing.T) {
	root := t.TempDir()
	_, err := FindRoot(root)
	if err == nil {
		t.Fatalf("expected error for missing prumo.json")
	}
}

func TestLoadManifest(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "prumo.json"), []byte(`{"version":3,"protocol":{"version":3},"project":{"name":"demo"}}`), 0644); err != nil {
		t.Fatalf("failed to create prumo.json: %v", err)
	}
	manifest, err := LoadManifest(root)
	if err != nil {
		t.Fatalf("expected manifest to load, got error: %v", err)
	}
	if manifest.Version != 3 {
		t.Fatalf("expected version 3, got %d", manifest.Version)
	}
}

func TestLoadManifestInvalid(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "prumo.json"), []byte(`{"version":1}`), 0644); err != nil {
		t.Fatalf("failed to create prumo.json: %v", err)
	}
	if _, err := LoadManifest(root); err == nil {
		t.Fatalf("expected error for legacy manifest version")
	}
	if _, err := LoadManifest(t.TempDir()); err == nil {
		t.Fatalf("expected error for missing manifest")
	}
}
