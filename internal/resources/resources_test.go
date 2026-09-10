package resources

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSchema(t *testing.T) {
	root := t.TempDir()
	schemasDir := filepath.Join(root, "schemas")
	if err := os.MkdirAll(schemasDir, 0755); err != nil {
		t.Fatalf("failed to create schemas dir: %v", err)
	}
	content := `{"$schema": "http://json-schema.org/draft-07/schema#"}`
	if err := os.WriteFile(filepath.Join(schemasDir, "prumo.schema.json"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}
	data, err := OpenSchema(root, "prumo.schema.json")
	if err != nil {
		t.Fatalf("expected schema to open, got error: %v", err)
	}
	if string(data) != content {
		t.Fatalf("expected %s, got %s", content, string(data))
	}
}

func TestFindResourcesDir(t *testing.T) {
	got := FindResourcesDir("/repo")
	expected := filepath.Join("/repo", "src", "prumo", "resources")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestOpenSchemaRejectsTraversal(t *testing.T) {
	if _, err := OpenSchema(t.TempDir(), "../secret.json"); err == nil {
		t.Fatalf("expected traversal path to be rejected")
	}
}
