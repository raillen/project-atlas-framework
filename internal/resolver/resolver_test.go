package resolver

import "testing"

func TestResolveBrasa(t *testing.T) {
	catalog, err := LoadCatalog("../..")
	if err != nil {
		t.Skipf("catalog missing: %v", err)
	}
	profile, err := LoadProfile("../../examples/brasa/project-profile.json")
	if err != nil {
		t.Skipf("profile missing: %v", err)
	}
	resolution := catalog.Resolve(profile)
	if len(resolution.Skills) == 0 {
		t.Fatalf("expected skills to be resolved")
	}
	seen := map[string]bool{}
	for _, skill := range resolution.Skills {
		seen[skill] = true
	}
	for _, expected := range []string{"clean-code"} {
		if !seen[expected] {
			t.Fatalf("expected skill %s in %v", expected, resolution.Skills)
		}
	}
}
