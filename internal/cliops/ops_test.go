package cliops

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	return root
}

func goldenResolution(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "conformance", "golden", "cli", "resolve-brasa-json.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(entry["stdout"].(string)), &value); err != nil {
		t.Fatalf("parse stdout: %v", err)
	}
	return value
}

func pythonResolution(t *testing.T, profile string) map[string]any {
	t.Helper()
	cmd := exec.Command("python3", "-c", "import json;from project_atlas.profile import load_profile;from project_atlas.resolver import resolve;import dataclasses;r=resolve(load_profile(__import__('pathlib').Path('"+profile+"')));print(json.dumps({'agents':r.agents,'skills':r.skills,'recipes':r.recipes,'reasons':r.reasons}))")
	cmd.Dir = repoRoot(t)
	cmd.Env = append(os.Environ(), "PYTHONPATH=src")
	out, err := cmd.Output()
	if err != nil {
		return goldenResolution(t)
	}
	var value map[string]any
	if err := json.Unmarshal(out, &value); err != nil {
		return goldenResolution(t)
	}
	return value
}

func TestResolveMatchesPythonBrasa(t *testing.T) {
	root := repoRoot(t)
	svc := New(root)
	got, err := svc.Resolve(filepath.Join(root, "examples", "brasa", "project-profile.json"))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	want := pythonResolution(t, filepath.Join(root, "examples", "brasa", "project-profile.json"))
	for _, key := range []string{"agents", "skills", "recipes"} {
		wantList, _ := want[key].([]any)
		gotList := []string{}
		switch key {
		case "agents":
			gotList = got.Agents
		case "skills":
			gotList = got.Skills
		case "recipes":
			gotList = got.Recipes
		}
		if len(wantList) != len(gotList) {
			t.Fatalf("%s length mismatch: python=%d go=%d", key, len(wantList), len(gotList))
		}
		for i := range wantList {
			if wantList[i] != gotList[i] {
				t.Fatalf("%s[%d] mismatch: python=%v go=%v", key, i, wantList[i], gotList[i])
			}
		}
	}
}

func TestInitValidateCompileSnapshot(t *testing.T) {
	root := repoRoot(t)
	svc := New(root)
	dir := t.TempDir()
	resolution, err := svc.Init(dir, filepath.Join(root, "examples", "brasa", "project-profile.json"))
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if len(resolution.Skills) == 0 {
		t.Fatalf("expected skills")
	}
	if errors := svc.Validate(dir); len(errors) != 0 {
		t.Fatalf("validate: %v", errors)
	}
	created, err := svc.Compile(dir, "generic")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("expected one generic artifact, got %v", created)
	}
	opencodeCreated, err := svc.Compile(dir, "opencode")
	if err != nil {
		t.Fatalf("compile opencode: %v", err)
	}
	if len(opencodeCreated) == 0 {
		t.Fatalf("expected opencode artifacts, got %v", opencodeCreated)
	}
	geminiCreated, err := svc.Compile(dir, "gemini")
	if err != nil {
		t.Fatalf("compile gemini: %v", err)
	}
	if len(geminiCreated) == 0 {
		t.Fatalf("expected gemini artifacts, got %v", geminiCreated)
	}
	snapshot := filepath.Join(dir, "snapshot.zip")
	if err := svc.Snapshot(dir, snapshot); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if _, err := os.Stat(snapshot); err != nil {
		t.Fatalf("snapshot missing: %v", err)
	}
	if err := svc.NewGoal(dir, "P00-G01", "Foundation", "P00", ""); err != nil {
		t.Fatalf("goal new: %v", err)
	}
	if _, err := svc.GoalState(dir, "P00-G01", "PLANNED", "test"); err != nil {
		t.Fatalf("goal state: %v", err)
	}
	plan := svc.ContextPlan(dir, "fix typo in label")
	if plan["profile"] != "small" {
		t.Fatalf("expected small profile, got %v", plan)
	}
	if _, err := svc.ReportAdd(dir, filepath.Join(root, "conformance", "fixtures", "task-report.json")); err != nil {
		t.Fatalf("report add: %v", err)
	}
	migrated, err := svc.Migrate(dir, true)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if migrated["to_version"] != 3 {
		t.Fatalf("expected to_version 3, got %v", migrated)
	}
	if findings, err := svc.Doctor(dir); err != nil || len(findings) != 0 {
		t.Fatalf("doctor: %v %v", findings, err)
	}
	if errors := svc.FrameworkCheck(); len(errors) != 0 {
		t.Fatalf("framework-check: %v", errors)
	}
}

func TestFrameworkCheckMatchesPython(t *testing.T) {
	root := repoRoot(t)
	svc := New(root)
	got := svc.FrameworkCheck()
	cmd := exec.Command("python3", "-c", "from project_atlas.validator import validate_framework;import json;print(json.dumps(validate_framework()))")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PYTHONPATH=src")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("python oracle unavailable: %v", err)
	}
	var want []string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("parse oracle: %v", err)
	}
	if len(want) != 0 || len(got) != 0 {
		t.Fatalf("framework mismatch: python=%v go=%v", want, got)
	}
}
