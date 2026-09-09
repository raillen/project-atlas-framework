package docengine

import (
	"path/filepath"
	"testing"
)

func TestRegistryProfilesComposeDeterministically(t *testing.T) {
	root, _ := filepath.Abs("../..")
	registry, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	contracts, unknown, err := registry.ResolveProfiles([]string{"cli", "compiler"}, map[string]bool{"cli": true, "compiler": true})
	if err != nil {
		t.Fatal(err)
	}
	if len(unknown) != 0 || len(contracts) == 0 {
		t.Fatalf("contracts=%d unknown=%v", len(contracts), unknown)
	}
	for i := 1; i < len(contracts); i++ {
		if contracts[i-1].ID > contracts[i].ID {
			t.Fatal("contracts not sorted")
		}
	}
}
func TestApplicability(t *testing.T) {
	contract := Contract{Applicability: map[string]any{"capabilities_any": []any{"desktop-gui"}}}
	if Applicable(contract, map[string]bool{"cli": true}) {
		t.Fatal("desktop contract applied to CLI")
	}
	if !Applicable(contract, map[string]bool{"desktop-gui": true}) {
		t.Fatal("desktop contract not applied")
	}
}
