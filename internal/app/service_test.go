package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/raillen/project-atlas-framework/internal/protocol"
)

func TestServiceVersion(t *testing.T) {
	svc := NewService()
	info := svc.Version(context.Background())
	if info.Version != protocol.CLIVersion {
		t.Fatalf("expected version %s, got %s", protocol.CLIVersion, info.Version)
	}
	if info.ProtocolVersion != protocol.ProtocolVersion {
		t.Fatalf("expected protocol %s, got %s", protocol.ProtocolVersion, info.ProtocolVersion)
	}
}

func TestServiceProjectRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "atlas.json"), []byte(`{"version":2}`), 0644); err != nil {
		t.Fatalf("failed to create atlas.json: %v", err)
	}
	svc := NewService()
	info, err := svc.ProjectRoot(context.Background(), root)
	if err != nil {
		t.Fatalf("expected project root, got error: %v", err)
	}
	if info.Manifest.Version != 2 {
		t.Fatalf("expected version 2, got %d", info.Manifest.Version)
	}
	if _, err := svc.ProjectRoot(context.Background(), t.TempDir()); err == nil {
		t.Fatalf("expected error for missing project")
	}
}
