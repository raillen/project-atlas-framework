package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }

func TestRunSessionAndContinuation(t *testing.T) {
	run := NewRun("R1", "G1", "T1")
	var err error
	run, err = run.Transition(RunPlanning)
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Transition(RunReady)
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Transition(RunRunning)
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession("S1", run.ID, "generic", "model-a")
	run = run.AttachSession(session)
	session = session.End("provider_limit")
	record := ContinuationFromRun(run, RepositoryState{Branch: "feat/g1", Revision: "abc", Dirty: true}, []string{"domain"}, []string{"cli"}, []string{"tests"}, []string{"implement CLI"})
	if record.Run != "R1" || record.Repository.Branch != "feat/g1" {
		t.Fatalf("record: %#v", record)
	}
	prompt := RenderPrompt(record)
	if len(prompt) == 0 {
		t.Fatal("empty prompt")
	}
	if session.Status != "ended" || len(run.Sessions) != 1 {
		t.Fatalf("session/run: %#v %#v", session, run)
	}
}
func TestRunRejectsInvalidTransition(t *testing.T) {
	run := NewRun("R1", "G1", "T1")
	if _, err := run.Transition(RunCompleted); err == nil {
		t.Fatal("expected invalid transition")
	}
}
func TestRunCancelResumeAndRepositoryInspection(t *testing.T) {
	run := NewRun("R1", "G1", "T1")
	var err error
	run, err = run.Transition(RunPlanning)
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Transition(RunReady)
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Transition(RunRunning)
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Transition(RunCancelled)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != RunCancelled || run.FinishedAt == "" {
		t.Fatalf("cancelled run: %#v", run)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for !fileExists(filepath.Join(root, "go.mod")) {
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("repository root not found")
		}
		root = parent
	}
	state := InspectRepository(root)
	if state.Revision == "" {
		t.Fatalf("repository state: %#v", state)
	}
}

func TestPendingSideEffectVisible(t *testing.T) {
	run := NewRun("R1", "G1", "T1")
	run.PendingSideEffects = []string{"remote:unknown"}
	record := ContinuationFromRun(run, RepositoryState{}, nil, nil, nil, []string{"reconcile side effect"})
	if len(record.PendingSideEffects) != 1 {
		t.Fatal("missing side effect")
	}
}
