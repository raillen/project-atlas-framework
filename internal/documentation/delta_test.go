package docengine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/raillen/prumo/internal/validation"
)

func writeDeltaFixture(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		"docs/contracts/builtin.json": `[{"id":"architecture.system","version":1,"role":"system-architecture"},{"id":"testing.strategy","version":1,"role":"testing-strategy"}]`,
		"docs/profiles/builtin.json":  `[{"id":"core-software","version":1,"contracts":["architecture.system","testing.strategy"]}]`,
		`docs/contracts/bindings.json`: `[
			{"contract_id":"architecture.system","sources":["docs/architecture/overview.md"],"ownership":"human","authority":"canonical"},
			{"contract_id":"testing.strategy","sources":["docs/development/testing-strategy.md"],"ownership":"human","authority":"canonical"}
		]`,
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewDeltaIsDeterministic(t *testing.T) {
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	impacts := []Impact{
		{ContractID: "testing.strategy", Documents: []string{"docs/development/testing-strategy.md"}},
		{ContractID: "architecture.system", Documents: []string{"docs/architecture/overview.md", "docs/architecture/dependency-rules.md"}},
	}
	first := NewDelta("G042", impacts, now)
	second := NewDelta("G042", []Impact{impacts[1], impacts[0]}, now)
	if first.ID != second.ID {
		t.Fatalf("deterministic deltas differ: %q and %q", first.ID, second.ID)
	}
	if first.Version != 1 || first.State != "proposed" {
		t.Fatalf("unexpected tracked delta: %#v", first)
	}
	if first.CreatedAt != "2026-09-09T12:00:00Z" || first.UpdatedAt != first.CreatedAt {
		t.Fatalf("unexpected timestamps: %#v", first)
	}
	if err := first.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDeltaLifecycleTransitions(t *testing.T) {
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	delta := NewDelta("G042", []Impact{{ContractID: "architecture.system", Documents: []string{"docs/architecture/overview.md"}}}, now)
	reviewed, err := delta.Transition("reviewed", []string{"EV-001"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Version != 2 || reviewed.CreatedAt != delta.CreatedAt {
		t.Fatalf("review must preserve identity and advance version: %#v", reviewed)
	}
	if _, err := reviewed.Transition("applied", nil, now); err == nil {
		t.Fatal("expected accepted to require an explicit transition before applied")
	}
	accepted, err := reviewed.Transition("accepted", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := accepted.Transition("proposed", nil, now); err == nil {
		t.Fatal("expected backward transitions to be rejected")
	}
	unevidenced := Delta{ID: "DD-0123456789ab", Version: 2, State: "accepted", Source: "G042", Contracts: []string{"architecture.system"}, Documents: []string{"docs/architecture/overview.md"}, Reason: "documentation impact analysis", Evidence: []string{}}
	if _, err := unevidenced.Transition("applied", nil, now); err == nil {
		t.Fatal("expected applied to require evidence")
	}
	applied, err := accepted.Transition("applied", []string{"EV-002"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if applied.State != "applied" || len(applied.Evidence) != 2 {
		t.Fatalf("unexpected applied delta: %#v", applied)
	}
}

func TestDeltaStorageRoundTrip(t *testing.T) {
	root := t.TempDir()
	writeDeltaFixture(t, root)
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	registry, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := LoadBindings(root)
	if err != nil {
		t.Fatal(err)
	}
	proposal := NewDelta("G042", AnalyzeImpacts(registry, bindings, []string{"docs/architecture/overview.md"}), now)
	if err := SaveDelta(root, proposal); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadDelta(root, proposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != proposal.ID || loaded.CreatedAt != proposal.CreatedAt {
		t.Fatalf("stored delta changed: %#v", loaded)
	}
	transitioned, err := TransitionDelta(root, proposal.ID, "reviewed", []string{"EV-001"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if transitioned.State != "reviewed" || transitioned.Version != 2 {
		t.Fatalf("unexpected transitioned delta: %#v", transitioned)
	}
	listed, err := ListDeltas(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != proposal.ID {
		t.Fatalf("unexpected delta list: %#v", listed)
	}
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".ai", "docs", "deltas", proposal.ID+".json")
	if errors := validation.ValidateFile(path, "documentation-delta.schema.json", filepath.Join(repoRoot, "schemas")); len(errors) != 0 {
		t.Fatalf("stored delta failed schema validation: %v", errors)
	}
}

func TestDeltaStorageRejectsCorruption(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".ai", "docs", "deltas", "DD-0123456789ab.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDelta(root, "DD-0123456789ab"); err == nil {
		t.Fatal("expected corrupted delta to fail")
	}
	if _, err := ListDeltas(root); err == nil {
		t.Fatal("expected delta listing to expose corruption")
	}
}

func TestDeltaConformanceFixture(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repoRoot, "conformance", "documentation", "valid_delta.json")
	if errors := validation.ValidateFile(path, "documentation-delta.schema.json", filepath.Join(repoRoot, "schemas")); len(errors) != 0 {
		t.Fatalf("delta fixture failed schema validation: %v", errors)
	}
}
