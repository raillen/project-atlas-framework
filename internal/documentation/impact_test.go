package docengine

import "testing"

func TestCoverageStates(t *testing.T) {
	states := []CoverageState{Missing, Partial, ImplementationReady, Verified, Stale, NotApplicable}
	seen := map[CoverageState]bool{}
	for _, state := range states {
		seen[state] = true
	}
	if len(seen) != 6 {
		t.Fatalf("expected six coverage states")
	}
}

func TestDeltaDeterministic(t *testing.T) {
	delta := MakeDelta("G042", []Impact{{ContractID: "b", Documents: []string{"z.md", "a.md"}}, {ContractID: "a", Documents: []string{"a.md"}}})
	if err := delta.Validate(); err != nil {
		t.Fatal(err)
	}
	if delta.Contracts[0] != "a" || delta.Documents[0] != "a.md" {
		t.Fatalf("not sorted: %#v", delta)
	}
}
