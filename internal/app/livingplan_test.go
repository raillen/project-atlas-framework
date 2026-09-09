package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/raillen/project-atlas-framework/internal/documentation"
	"github.com/raillen/project-atlas-framework/internal/planning"
)

func writeLivingPlanFixture(t *testing.T, root string) {
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

func sampleAcceptedDecisions() []planning.DecisionProposal {
	return []planning.DecisionProposal{
		{
			ID:             "DP-001",
			Statement:      "Atlas must expose a provider-neutral planning review state",
			Classification: "explicit-decision",
			Scope:          "planning",
			Actor:          "owner",
			Source:         "User",
			Authority:      "user-decision",
			Confidence:     planning.ConfidenceUnknown,
			Status:         planning.StatusAccepted,
			Affected:       []string{"architecture.system"},
			Rationale:      "needed for the M6 living plan loop",
		},
		{
			ID:             "DP-003",
			Statement:      "Adopt Python 3.12 for runtime tooling",
			Classification: "explicit-decision",
			Scope:          "tooling",
			Actor:          "owner",
			Source:         "User",
			Authority:      "project-decision",
			Confidence:     planning.ConfidenceUnknown,
			Status:         planning.StatusAccepted,
			Affected:       []string{"testing.strategy"},
			Rationale:      "unblocks CI conformance",
		},
	}
}

func TestImpactsFromDecisionsOnlyAccepted(t *testing.T) {
	decisions := sampleAcceptedDecisions()
	decisions = append(decisions, planning.DecisionProposal{
		ID: "DP-S", Statement: "suggestion", Classification: "agent-suggestion", Status: planning.StatusProposed, Affected: []string{"architecture.system"},
	})
	bound := []docengine.Binding{
		{ContractID: "architecture.system", Sources: []string{"docs/architecture/overview.md"}},
		{ContractID: "testing.strategy", Sources: []string{"docs/development/testing-strategy.md"}},
	}
	impacts := ImpactsFromDecisions(bound, decisions)
	if len(impacts) != 2 {
		t.Fatalf("expected 2 impacts, got %d", len(impacts))
	}
	if impacts[0].ContractID != "architecture.system" || impacts[0].Reason != "living plan decision DP-001" {
		t.Fatalf("unexpected first impact: %#v", impacts[0])
	}
	if len(impacts[0].Documents) != 1 || impacts[0].Documents[0] != "docs/architecture/overview.md" {
		t.Fatalf("bound documents not carried: %#v", impacts[0])
	}
	for _, impact := range impacts {
		if impact.ContractID == "architecture.system" && impact.Reason == "living plan decision DP-S" {
			t.Fatal("agent suggestion must not drive documentation impact")
		}
	}
}

func TestImpactsFromDecisionsDeduplicatesContracts(t *testing.T) {
	decisions := sampleAcceptedDecisions()
	decisions = append(decisions, planning.DecisionProposal{
		ID: "DP-009", Statement: "same contract again", Classification: "explicit-decision", Actor: "owner",
		Source: "User", Status: planning.StatusAccepted, Affected: []string{"architecture.system"},
	})
	impacts := ImpactsFromDecisions(nil, decisions)
	if len(impacts) != 2 {
		t.Fatalf("expected 2 deduplicated impacts, got %d", len(impacts))
	}
}

func TestPlanDeltaProducesProposal(t *testing.T) {
	root := t.TempDir()
	writeLivingPlanFixture(t, root)
	loop := LivingPlanLoop{Root: root}
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)

	delta, err := loop.PlanDelta("G042", sampleAcceptedDecisions(), now)
	if err != nil {
		t.Fatal(err)
	}
	if delta.State != "proposed" {
		t.Fatalf("expected proposed delta, got %s", delta.State)
	}
	if delta.Version != 1 {
		t.Fatalf("expected version 1, got %d", delta.Version)
	}
	if len(delta.Contracts) != 2 {
		t.Fatalf("expected 2 affected contracts, got %v", delta.Contracts)
	}
	if err := delta.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestApplyGovernanceReachesApplied(t *testing.T) {
	root := t.TempDir()
	writeLivingPlanFixture(t, root)
	loop := LivingPlanLoop{Root: root}
	now := time.Date(2026, time.September, 9, 12, 30, 0, 0, time.UTC)

	proposal, err := loop.PlanDelta("G042", sampleAcceptedDecisions(), now)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := loop.ApplyGovernance(proposal, []string{"DP-001", "DP-003"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if applied.State != "applied" {
		t.Fatalf("expected applied state, got %s", applied.State)
	}
	if len(applied.Evidence) == 0 {
		t.Fatal("applied delta must carry decision evidence")
	}
	if err := applied.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRunProducesCoherentFeedback(t *testing.T) {
	root := t.TempDir()
	writeLivingPlanFixture(t, root)
	loop := LivingPlanLoop{Root: root}
	now := time.Date(2026, time.September, 9, 13, 0, 0, 0, time.UTC)

	report, err := loop.Run("G042", sampleAcceptedDecisions(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Impacts) != 2 {
		t.Fatalf("expected 2 impacts, got %d", len(report.Impacts))
	}
	if report.Delta.State != "applied" {
		t.Fatalf("expected applied delta, got %s", report.Delta.State)
	}
	if report.Readiness.Goal != "G042" {
		t.Fatalf("readiness goal mismatch: %#v", report.Readiness)
	}
	if err := ValidateFeedback(report, sampleAcceptedDecisions()); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsEmptyAcceptedDecision(t *testing.T) {
	root := t.TempDir()
	writeLivingPlanFixture(t, root)
	loop := LivingPlanLoop{Root: root}
	now := time.Date(2026, time.September, 9, 13, 0, 0, 0, time.UTC)

	report, err := loop.Run("G042", []planning.DecisionProposal{
		{ID: "DP-S", Statement: "suggestion", Classification: "agent-suggestion", Status: planning.StatusProposed},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Impacts) != 0 {
		t.Fatalf("suggestions must not produce impacts, got %d", len(report.Impacts))
	}
	if err := ValidateFeedback(report, nil); err != nil {
		t.Fatal(err)
	}
}
