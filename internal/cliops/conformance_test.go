package cliops

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func canonicalJSON(t *testing.T, path string) any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return normalizeTimestamps(value)
}

func normalizeTimestamps(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, item := range typed {
			if key == "updated_at" || key == "locked_at" || key == "approved_at" || key == "Last updated" {
				out[key] = "TIMESTAMP"
				continue
			}
			out[key] = normalizeTimestamps(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = normalizeTimestamps(item)
		}
		return out
	default:
		return value
	}
}

func TestConformanceInitMatchesPython(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "src", "prumo", "__init__.py")); os.IsNotExist(err) {
		t.Skip("python oracle retired: Python implementation removed from src/")
	}
	pyDir := t.TempDir()
	goDir := t.TempDir()
	profile := filepath.Join(root, "examples", "brasa", "project-profile.json")
	pyCmd := exec.Command("python3", "-m", "prumo", "init", pyDir, "--profile", profile, "--non-interactive")
	pyCmd.Dir = root
	pyCmd.Env = append(os.Environ(), "PYTHONPATH=src")
	if out, err := pyCmd.CombinedOutput(); err != nil {
		t.Skipf("python oracle retired or unavailable: %v %s", err, out)
	}
	svc := New(root)
	if _, err := svc.Init(goDir, profile); err != nil {
		t.Fatalf("go init: %v", err)
	}
	compare := []string{"prumo.json", ".ai/agents/manifest.json", ".ai/skills/manifest.json", ".ai/recipes/manifest.json", ".ai/orchestration/model-policy.json", ".ai/orchestration/orchestrator.json", ".ai/orchestration/fallbacks.json", ".ai/orchestration/model-scorecard.json"}
	for _, rel := range compare {
		want := canonicalJSON(t, filepath.Join(pyDir, rel))
		got := canonicalJSON(t, filepath.Join(goDir, rel))
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("mismatch in %s", rel)
		}
	}
}

func TestConformanceCompileMatchesPython(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "src", "prumo", "__init__.py")); os.IsNotExist(err) {
		t.Skip("python oracle retired: Python implementation removed from src/")
	}
	pyDir := t.TempDir()
	goDir := t.TempDir()
	profile := filepath.Join(root, "examples", "brasa", "project-profile.json")
	pyInit := exec.Command("python3", "-m", "prumo", "init", pyDir, "--profile", profile, "--non-interactive")
	pyInit.Dir = root
	pyInit.Env = append(os.Environ(), "PYTHONPATH=src")
	if out, err := pyInit.CombinedOutput(); err != nil {
		t.Skipf("python oracle retired or unavailable: %v %s", err, out)
	}
	svc := New(root)
	if _, err := svc.Init(goDir, profile); err != nil {
		t.Fatalf("go init: %v", err)
	}
	for _, target := range []string{"generic", "chatgpt", "claude", "kimi", "codex", "claude-code", "traycer"} {
		pyCmd := exec.Command("python3", "-m", "prumo", "compile", "--target", target, "--path", pyDir)
		pyCmd.Dir = root
		pyCmd.Env = append(os.Environ(), "PYTHONPATH=src")
		if out, err := pyCmd.CombinedOutput(); err != nil {
			t.Fatalf("python compile %s: %v %s", target, err, out)
		}
		created, err := svc.Compile(goDir, target)
		if err != nil {
			t.Fatalf("go compile %s: %v", target, err)
		}
		for _, path := range created {
			rel, err := filepath.Rel(goDir, path)
			if err != nil {
				t.Fatalf("rel: %v", err)
			}
			pyPath := filepath.Join(pyDir, rel)
			want, err := os.ReadFile(pyPath)
			if err != nil {
				t.Fatalf("python artifact missing %s: %v", rel, err)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("go artifact missing %s: %v", rel, err)
			}
			if string(want) != string(got) {
				t.Fatalf("artifact mismatch: %s", rel)
			}
		}
	}
}
