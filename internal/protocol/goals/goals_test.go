package goals

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestComputeDigestMatchesPython(t *testing.T) {
	data, err := os.ReadFile("../../../conformance/goals/valid_locked_goal.json")
	if err != nil {
		t.Skipf("conformance fixture missing: %v", err)
	}
	var goal Goal
	if err := json.Unmarshal(data, &goal); err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}
	lock, _ := goal["lock"].(map[string]any)
	if ComputeDigest(goal) != lock["digest"] {
		t.Fatalf("digest mismatch: expected %v got %v", lock["digest"], ComputeDigest(goal))
	}
	if valid, _ := VerifyLock(goal); !valid {
		t.Fatalf("expected valid lock")
	}
}

func TestTransitionAndAmend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "P00-G01.goal.json")
	goal := NewGoal("P00-G01", "Foundation", "P00", "Ship foundation.")
	data, _ := json.Marshal(goal)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := TransitionGoal(path, "PLANNED", ""); err != nil {
		t.Fatalf("transition: %v", err)
	}
	if _, err := TransitionGoal(path, "LOCKED", ""); err != nil {
		t.Fatalf("lock: %v", err)
	}
	if _, err := TransitionGoal(path, "DONE", ""); err == nil {
		t.Fatalf("expected DONE without evidence to fail")
	}
	if _, err := AmendGoal(path, map[string]any{"changes": map[string]any{"objective": "Updated."}}); err != nil {
		t.Fatalf("amend: %v", err)
	}
}
