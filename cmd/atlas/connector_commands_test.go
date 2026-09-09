package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConnectorCommands(t *testing.T) {
	// Setup isolated test environment
	tmpHome := t.TempDir()
	tmpProject := t.TempDir()
	t.Setenv("ATLAS_HOME", tmpHome)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(tmpProject); err != nil {
		t.Fatalf("failed to chdir to tmpProject: %v", err)
	}

	// 1. List connectors
	if code := runConnector(false, []string{"list"}); code != exitOK {
		t.Fatalf("runConnector list expected %d, got %d", exitOK, code)
	}
	if code := runConnector(true, []string{"list"}); code != exitOK {
		t.Fatalf("runConnector --json list expected %d, got %d", exitOK, code)
	}

	// 2. Validate before install -> should fail
	if code := runConnector(false, []string{"validate", "opencode"}); code == exitOK {
		t.Fatalf("runConnector validate expected failure before install, got %d", code)
	}

	// 3. Install opencode
	if code := runConnector(false, []string{"install", "opencode"}); code != exitOK {
		t.Fatalf("runConnector install opencode expected %d, got %d", exitOK, code)
	}

	// Verify .opencode files exist
	opencodeConfig := filepath.Join(tmpProject, ".opencode", "opencode.json")
	if _, err := os.Stat(opencodeConfig); err != nil {
		t.Fatalf("expected opencode.json at %s", opencodeConfig)
	}

	// 4. Validate after install -> should succeed
	if code := runConnector(false, []string{"validate", "opencode"}); code != exitOK {
		t.Fatalf("runConnector validate after install expected %d, got %d", exitOK, code)
	}
	if code := runConnector(true, []string{"validate", "opencode"}); code != exitOK {
		t.Fatalf("runConnector --json validate after install expected %d, got %d", exitOK, code)
	}

	// 5. Uninstall opencode
	if code := runConnector(false, []string{"uninstall", "opencode"}); code != exitOK {
		t.Fatalf("runConnector uninstall opencode expected %d, got %d", exitOK, code)
	}

	// 6. Test backward-compatible 'atlas install connector opencode'
	if code := runInstall(false, tmpHome, []string{"connector", "opencode"}); code != exitOK {
		t.Fatalf("runInstall connector opencode expected %d, got %d", exitOK, code)
	}
	if code := runConnector(false, []string{"validate", "opencode"}); code != exitOK {
		t.Fatalf("runConnector validate after install connector expected %d, got %d", exitOK, code)
	}
	if code := runConnector(false, []string{"uninstall", "opencode"}); code != exitOK {
		t.Fatalf("runConnector final uninstall expected %d, got %d", exitOK, code)
	}
}
