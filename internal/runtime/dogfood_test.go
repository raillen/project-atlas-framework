package runtime

import (
	"strings"
	"testing"
)

func TestFreshExecutorContinuationDogfood(t *testing.T) {
	run := NewRun("R-DOGFOOD", "G-DOGFOOD", "T-DOGFOOD")
	run, _ = run.Transition(RunPlanning)
	run, _ = run.Transition(RunReady)
	run, _ = run.Transition(RunRunning)
	run.Evidence = []string{"EV-DOMAIN"}
	record := ContinuationFromRun(run, RepositoryState{Branch: "feat/dogfood", Revision: "abc123", Dirty: true}, []string{"domain implementation"}, []string{"CLI continuation"}, []string{"integration tests"}, []string{"implement CLI", "run tests"})
	prompt := RenderPrompt(record)
	if !strings.Contains(prompt, "Run: R-DOGFOOD") || !strings.Contains(prompt, "domain implementation") || !strings.Contains(prompt, "run tests") {
		t.Fatalf("continuation prompt lost state: %s", prompt)
	}
	if strings.Contains(prompt, "transcript") || strings.Contains(prompt, "chain-of-thought") {
		t.Fatalf("continuation leaked private session state")
	}
}
