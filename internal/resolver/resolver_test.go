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

func TestResolveSystemsSkills(t *testing.T) {
	catalog, err := LoadCatalog("../..")
	if err != nil {
		t.Fatalf("catalog missing: %v", err)
	}

	tests := []struct {
		name          string
		stack         []string
		expectedSkill string
	}{
		{"C++ stack", []string{"cpp", "cmake"}, "lang-cpp"},
		{"C stack", []string{"c", "gcc"}, "lang-c"},
		{"Zig stack", []string{"zig"}, "lang-zig"},
		{"D stack", []string{"dlang"}, "lang-d"},
		{"Odin stack", []string{"odin"}, "lang-odin"},
		{"C3 stack", []string{"c3"}, "lang-c3"},
		{"Assembly stack", []string{"asm", "x86_64"}, "lang-asm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := Profile{
				Raw: map[string]any{
					"stack": tt.stack,
				},
			}
			resolution := catalog.Resolve(profile)
			found := false
			for _, s := range resolution.Skills {
				if s == tt.expectedSkill {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected skill %s to resolve for stack %v, got %v", tt.expectedSkill, tt.stack, resolution.Skills)
			}
		})
	}
}
