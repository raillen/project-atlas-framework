package docengine

import (
	"path/filepath"
	"testing"
)

func TestPrumoAuditAndReadiness(t *testing.T) {
	root, _ := filepath.Abs("../..")
	audit, err := Audit(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.ApplicableContracts) == 0 {
		t.Fatal("expected applicable contracts")
	}
	readiness, err := Readiness(root, "M5")
	if err != nil {
		t.Fatal(err)
	}
	if len(readiness.Coverage) == 0 {
		t.Fatal("expected coverage")
	}
}

func TestVerifiedCoverageRequiresEvidence(t *testing.T) {
	contract := Contract{ID: "verified", RequiredKnowledge: []string{"architecture"}, EvidenceRequirements: []string{"review"}}
	withoutEvidence := evaluate("../..", contract, Binding{Sources: []string{"docs/architecture/overview.md"}})
	if withoutEvidence.State != ImplementationReady {
		t.Fatalf("unexpected state: %s", withoutEvidence.State)
	}
	withEvidence := evaluate("../..", contract, Binding{Sources: []string{"docs/architecture/overview.md"}, Evidence: []string{"review-result"}})
	if withEvidence.State != Verified {
		t.Fatalf("expected verified, got %s", withEvidence.State)
	}
}
