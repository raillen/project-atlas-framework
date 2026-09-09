package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raillen/project-atlas-framework/internal/protocol"
)

func TestRunAdoptHumanReport(t *testing.T) {
	root := t.TempDir()
	writeTestFileForCLI(t, root, "go.mod", "module example.com/clitest\n\ngo 1.22\n")
	writeTestFileForCLI(t, root, "cmd/tool/main.go", "package main\n\nfunc main() {}\n")
	writeTestFileForCLI(t, root, "README.md", "# Tool\n\nA CLI demonstration.\n")

	code, out := captureOutput(func() int {
		return run([]string{"adopt", "--path", root, "--audit-only"})
	})

	if code != exitOK {
		t.Fatalf("expected exitOK (%d), got %d; output:\n%s", exitOK, code, out)
	}

	if !strings.Contains(out, "PROJECT ATLAS — REPOSITORY ADOPTION REPORT") {
		t.Errorf("expected adoption report header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1. Classification") {
		t.Errorf("expected classification section in output")
	}
	if !strings.Contains(out, "2. Recommended Profiles & Capabilities") {
		t.Errorf("expected profiles section in output")
	}
}

func TestRunAdoptJSONReport(t *testing.T) {
	root := t.TempDir()
	writeTestFileForCLI(t, root, "go.mod", "module example.com/clitest\n\ngo 1.22\n")
	writeTestFileForCLI(t, root, "cmd/tool/main.go", "package main\n\nfunc main() {}\n")
	writeTestFileForCLI(t, root, "README.md", "# Tool\n\nA CLI demonstration.\n")

	code, out := captureOutput(func() int {
		return run([]string{"--json", "adopt", "--path", root, "--audit-only"})
	})

	if code != exitOK {
		t.Fatalf("expected exitOK (%d), got %d; output:\n%s", exitOK, code, out)
	}

	var env protocol.Envelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("failed to parse JSON envelope: %v; raw output:\n%s", err, out)
	}
	if !env.Ok {
		t.Fatalf("expected envelope ok=true, got false")
	}

	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map in envelope, got %T", env.Data)
	}
	if dataMap["version"] == nil {
		t.Errorf("expected version field in report data")
	}
	if dataMap["repository"] == nil {
		t.Errorf("expected repository field in report data")
	}
	if dataMap["classification"] == nil {
		t.Errorf("expected classification field in report data")
	}
}

func TestRunAdoptStrictFailureOnUnknown(t *testing.T) {
	root := t.TempDir()
	// Write a source file with no known project archetype (triggers unknown app-type)
	writeTestFileForCLI(t, root, "script.sh", "#!/bin/sh\necho mystery\n")

	// Test non-JSON mode
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	code := run([]string{"adopt", "--path", root, "--strict"})
	_ = w.Close()
	os.Stderr = oldStderr
	var errBuf strings.Builder
	var buf [1024]byte
	n, _ := r.Read(buf[:])
	errBuf.Write(buf[:n])

	if code != exitValidation {
		t.Fatalf("expected exitValidation (%d), got %d", exitValidation, code)
	}
	if !strings.Contains(errBuf.String(), "strict adoption check failed") {
		t.Errorf("expected strict failure message in stderr, got:\n%s", errBuf.String())
	}

	// Test JSON mode
	codeJSON, outJSON := captureOutput(func() int {
		return run([]string{"--json", "adopt", "--path", root, "--strict"})
	})
	if codeJSON != exitValidation {
		t.Fatalf("expected exitValidation (%d), got %d", exitValidation, codeJSON)
	}
	var env protocol.Envelope
	if err := json.Unmarshal([]byte(outJSON), &env); err != nil {
		t.Fatalf("failed to parse JSON envelope: %v", err)
	}
	if env.Ok {
		t.Errorf("expected envelope ok=false for strict failure")
	}
	if len(env.Diagnostics) == 0 || env.Diagnostics[0].Code != "adoption_strict_failure" {
		t.Errorf("expected diagnostic code 'adoption_strict_failure', got %v", env.Diagnostics)
	}
}

func TestRunAdoptNonInteractive(t *testing.T) {
	root := t.TempDir()
	writeTestFileForCLI(t, root, "go.mod", "module example.com/clitest\n\ngo 1.22\n")
	writeTestFileForCLI(t, root, "cmd/tool/main.go", "package main\n\nfunc main() {}\n")

	code, out := captureOutput(func() int {
		return run([]string{"adopt", "--path", root, "--non-interactive"})
	})

	if code != exitOK {
		t.Fatalf("expected exitOK (%d), got %d; output:\n%s", exitOK, code, out)
	}
	if !strings.Contains(out, "Non-interactive pass") && !strings.Contains(out, "PROJECT ATLAS") {
		t.Errorf("expected adoption output in non-interactive mode, got:\n%s", out)
	}
}

func writeTestFileForCLI(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
